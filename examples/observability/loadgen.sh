#!/usr/bin/env bash
# Load generator for the observability sample stack.
#
# Usage (local):
#   chmod +x loadgen.sh && ./loadgen.sh
#   ./loadgen.sh --url http://localhost:18080 --rate 5 --duration 60
#
# Via docker compose:
#   docker compose --profile load up loadgen
#
# Environment variables (alternative to flags):
#   API_URL, RATE, DURATION

set -uo pipefail

API_URL="${API_URL:-http://localhost:18080}"
RATE="${RATE:-2}"
DURATION="${DURATION:-0}"   # seconds; 0 = run until Ctrl-C

while [[ $# -gt 0 ]]; do
  case $1 in
    --url)      API_URL="$2";    shift 2 ;;
    --rate)     RATE="$2";       shift 2 ;;
    --duration) DURATION="$2";   shift 2 ;;
    *) printf "unknown flag: %s\n" "$1" >&2; exit 1 ;;
  esac
done

SLEEP=$(awk "BEGIN{printf \"%.3f\", 1/$RATE}")

GRN='\033[0;32m' RED='\033[0;31m' CYN='\033[0;36m' DIM='\033[2m' RST='\033[0m'

total=0; oks=0; fails=0

summary() {
  printf "\n${CYN}done${RST}  total=%d  ok=${GRN}%d${RST}  errors=${RED}%d${RST}\n" "$total" "$oks" "$fails"
}
trap summary EXIT INT TERM

# request METHOD PATH [BODY]
# Prints a timestamped line and updates counters.
request() {
  local method="$1" path="$2" body="${3:-}"
  local curl_args=(-s -o /dev/null -w "%{http_code}" -X "$method" -H "Accept: application/json")
  [[ -n "$body" ]] && curl_args+=(-H "Content-Type: application/json" -d "$body")

  local code
  code=$(curl "${curl_args[@]}" "${API_URL}${path}" 2>/dev/null || printf "000")

  total=$(( total + 1 ))
  local ts; ts=$(date +%H:%M:%S)

  if [[ "$code" =~ ^[23] ]]; then
    printf "${GRN}%s  %-5s  %-35s  %s${RST}\n" "$ts" "$method" "$path" "$code"
    oks=$(( oks + 1 ))
  else
    printf "${RED}%s  %-5s  %-35s  %s${RST}\n" "$ts" "$method" "$path" "$code"
    fails=$(( fails + 1 ))
  fi
}

# Fixed set of item names used across POST requests so the store grows naturally.
NAMES=("Sprocket" "Cog" "Flange" "Gasket" "Bushing" "Valve" "Piston" "Lever" "Bearing" "Cam")
VALID_IDS=(1 2 3)

run_one() {
  local n=$(( RANDOM % 10 ))
  case $n in
    0)
      request GET /health
      ;;
    1|2)
      request GET /api/v1/items
      ;;
    3|4)
      local id="${VALID_IDS[$(( RANDOM % ${#VALID_IDS[@]} ))]}"
      request GET "/api/v1/items/$id"
      ;;
    5)
      # Intentional 404 — shows error spans in Tempo.
      request GET /api/v1/items/999
      ;;
    6)
      local name="${NAMES[$(( RANDOM % ${#NAMES[@]} ))]}"
      request POST /api/v1/items "{\"name\":\"$name\"}"
      ;;
    7)
      # Intentional 400 — shows validation error in traces.
      request POST /api/v1/items '{"name":""}'
      ;;
    8|9)
      local id="${VALID_IDS[$(( RANDOM % ${#VALID_IDS[@]} ))]}"
      request GET "/api/v1/items/$id"
      ;;
  esac
}

printf "${CYN}loadgen${RST}  url=${DIM}%s${RST}  rate=${DIM}%s req/s${RST}\n" "$API_URL" "$RATE"
[[ "$DURATION" -gt 0 ]] && printf "         duration=${DIM}%ss${RST}\n" "$DURATION"
printf "${DIM}%-8s  %-5s  %-35s  status${RST}\n\n" "time" "verb" "path"

if (( DURATION > 0 )); then
  end=$(( $(date +%s) + DURATION ))
  while (( $(date +%s) < end )); do
    run_one
    sleep "$SLEEP"
  done
else
  while true; do
    run_one
    sleep "$SLEEP"
  done
fi
