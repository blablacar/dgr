package bats
import "io/ioutil"

const (
	bats=`#!/usr/bin/env bash
set -e

version() {
  echo "Bats 0.4.0"
}

usage() {
  version
  echo "Usage: bats [-c] [-p | -t] <test> [<test> ...]"
}

help() {
  usage
  echo
  echo "  <test> is the path to a Bats test file, or the path to a directory"
  echo "  containing Bats test files."
  echo
  echo "  -c, --count    Count the number of test cases without running any tests"
  echo "  -h, --help     Display this help message"
  echo "  -p, --pretty   Show results in pretty format (default for terminals)"
  echo "  -t, --tap      Show results in TAP format"
  echo "  -v, --version  Display the version number"
  echo
  echo "  For more information, see https://github.com/sstephenson/bats"
  echo
}

resolve_link() {
  $(type -p greadlink readlink | head -1) "$1"
}

abs_dirname() {
  local cwd="$(pwd)"
  local path="$1"

  while [ -n "$path" ]; do
    cd "${path%/*}"
    local name="${path##*/}"
    path="$(resolve_link "$name" || true)"
  done

  pwd
  cd "$cwd"
}

expand_path() {
  { cd "$(dirname "$1")" 2>/dev/null
    local dirname="$PWD"
    cd "$OLDPWD"
    echo "$dirname/$(basename "$1")"
  } || echo "$1"
}

BATS_LIBEXEC="$(abs_dirname "$0")"
export BATS_PREFIX="$(abs_dirname "$BATS_LIBEXEC")"
export BATS_CWD="$(abs_dirname .)"
export PATH="$BATS_LIBEXEC:$PATH"

options=()
arguments=()
for arg in "$@"; do
  if [ "${arg:0:1}" = "-" ]; then
    if [ "${arg:1:1}" = "-" ]; then
      options[${#options[*]}]="${arg:2}"
    else
      index=1
      while option="${arg:$index:1}"; do
        [ -n "$option" ] || break
        options[${#options[*]}]="$option"
        let index+=1
      done
    fi
  else
    arguments[${#arguments[*]}]="$arg"
  fi
done

unset count_flag pretty
[ -t 0 ] && [ -t 1 ] && pretty="1"
[ -n "$CI" ] && pretty=""

for option in "${options[@]}"; do
  case "$option" in
  "h" | "help" )
    help
    exit 0
    ;;
  "v" | "version" )
    version
    exit 0
    ;;
  "c" | "count" )
    count_flag="-c"
    ;;
  "t" | "tap" )
    pretty=""
    ;;
  "p" | "pretty" )
    pretty="1"
    ;;
  * )
    usage >&2
    exit 1
    ;;
  esac
done

if [ "${#arguments[@]}" -eq 0 ]; then
  usage >&2
  exit 1
fi

filenames=()
for filename in "${arguments[@]}"; do
  if [ -d "$filename" ]; then
    shopt -s nullglob
    for suite_filename in "$(expand_path "$filename")"/*.bats; do
      filenames["${#filenames[@]}"]="$suite_filename"
    done
    shopt -u nullglob
  else
    filenames["${#filenames[@]}"]="$(expand_path "$filename")"
  fi
done

if [ "${#filenames[@]}" -eq 1 ]; then
  command="bats-exec-test"
else
  command="bats-exec-suite"
fi

if [ -n "$pretty" ]; then
  extended_syntax_flag="-x"
  formatter="bats-format-tap-stream"
else
  extended_syntax_flag=""
  formatter="cat"
fi

set -o pipefail execfail
exec "$command" $count_flag $extended_syntax_flag "${filenames[@]}" | "$formatter"`
	bats_exec_suite = `#!/usr/bin/env bash
set -e

count_only_flag=""
if [ "$1" = "-c" ]; then
  count_only_flag=1
  shift
fi

extended_syntax_flag=""
if [ "$1" = "-x" ]; then
  extended_syntax_flag="-x"
  shift
fi

trap "kill 0; exit 1" int

count=0
for filename in "$@"; do
  let count+="$(bats-exec-test -c "$filename")"
done

if [ -n "$count_only_flag" ]; then
  echo "$count"
  exit
fi

echo "1..$count"
status=0
offset=0
for filename in "$@"; do
  index=0
  {
    IFS= read -r # 1..n
    while IFS= read -r line; do
      case "$line" in
      "begin "* )
        let index+=1
        echo "${line/ $index / $(($offset + $index)) }"
        ;;
      "ok "* | "not ok "* )
        [ -n "$extended_syntax_flag" ] || let index+=1
        echo "${line/ $index / $(($offset + $index)) }"
        [ "${line:0:6}" != "not ok" ] || status=1
        ;;
      * )
        echo "$line"
        ;;
      esac
    done
  } < <( bats-exec-test $extended_syntax_flag "$filename" )
  offset=$(($offset + $index))
done

exit "$status"`
	bats_exec_test=`#!/usr/bin/env bash
set -e
set -E
set -T

BATS_COUNT_ONLY=""
if [ "$1" = "-c" ]; then
  BATS_COUNT_ONLY=1
  shift
fi

BATS_EXTENDED_SYNTAX=""
if [ "$1" = "-x" ]; then
  BATS_EXTENDED_SYNTAX="$1"
  shift
fi

BATS_TEST_FILENAME="$1"
if [ -z "$BATS_TEST_FILENAME" ]; then
  echo "usage: bats-exec <filename>" >&2
  exit 1
elif [ ! -f "$BATS_TEST_FILENAME" ]; then
  echo "bats: $BATS_TEST_FILENAME does not exist" >&2
  exit 1
else
  shift
fi

BATS_TEST_DIRNAME="$(dirname "$BATS_TEST_FILENAME")"
BATS_TEST_NAMES=()

load() {
  local name="$1"
  local filename

  if [ "${name:0:1}" = "/" ]; then
    filename="${name}"
  else
    filename="$BATS_TEST_DIRNAME/${name}.bash"
  fi

  [ -f "$filename" ] || {
    echo "bats: $filename does not exist" >&2
    exit 1
  }

  source "${filename}"
}

run() {
  local e E T oldIFS
  [[ ! "$-" =~ e ]] || e=1
  [[ ! "$-" =~ E ]] || E=1
  [[ ! "$-" =~ T ]] || T=1
  set +e
  set +E
  set +T
  output="$("$@" 2>&1)"
  status="$?"
  oldIFS=$IFS
  IFS=$'\n' lines=($output)
  [ -z "$e" ] || set -e
  [ -z "$E" ] || set -E
  [ -z "$T" ] || set -T
  IFS=$oldIFS
}

setup() {
  true
}

teardown() {
  true
}

skip() {
  BATS_TEST_SKIPPED=${1:-1}
  BATS_TEST_COMPLETED=1
  exit 0
}

bats_test_begin() {
  BATS_TEST_DESCRIPTION="$1"
  if [ -n "$BATS_EXTENDED_SYNTAX" ]; then
    echo "begin $BATS_TEST_NUMBER $BATS_TEST_DESCRIPTION" >&3
  fi
  setup
}

bats_test_function() {
  local test_name="$1"
  BATS_TEST_NAMES["${#BATS_TEST_NAMES[@]}"]="$test_name"
}

bats_capture_stack_trace() {
  BATS_PREVIOUS_STACK_TRACE=( "${BATS_CURRENT_STACK_TRACE[@]}" )
  BATS_CURRENT_STACK_TRACE=()

  local test_pattern=" $BATS_TEST_NAME $BATS_TEST_SOURCE"
  local setup_pattern=" setup $BATS_TEST_SOURCE"
  local teardown_pattern=" teardown $BATS_TEST_SOURCE"

  local frame
  local index=1

  while frame="$(caller "$index")"; do
    BATS_CURRENT_STACK_TRACE["${#BATS_CURRENT_STACK_TRACE[@]}"]="$frame"
    if [[ "$frame" = *"$test_pattern"     || \
          "$frame" = *"$setup_pattern"    || \
          "$frame" = *"$teardown_pattern" ]]; then
      break
    else
      let index+=1
    fi
  done

  BATS_SOURCE="$(bats_frame_filename "${BATS_CURRENT_STACK_TRACE[0]}")"
  BATS_LINENO="$(bats_frame_lineno "${BATS_CURRENT_STACK_TRACE[0]}")"
}

bats_print_stack_trace() {
  local frame
  local index=1
  local count="${#@}"

  for frame in "$@"; do
    local filename="$(bats_trim_filename "$(bats_frame_filename "$frame")")"
    local lineno="$(bats_frame_lineno "$frame")"

    if [ $index -eq 1 ]; then
      echo -n "# ("
    else
      echo -n "#  "
    fi

    local fn="$(bats_frame_function "$frame")"
    if [ "$fn" != "$BATS_TEST_NAME" ]; then
      echo -n "from function '$fn' "
fi

if [ $index -eq $count ]; then
echo "in test file $filename, line $lineno)"
else
echo "in file $filename, line $lineno,"
fi

let index+=1
done
}

bats_print_failed_command() {
local frame="$1"
local status="$2"
local filename="$(bats_frame_filename "$frame")"
local lineno="$(bats_frame_lineno "$frame")"

local failed_line="$(bats_extract_line "$filename" "$lineno")"
local failed_command="$(bats_strip_string "$failed_line")"
echo -n "#   '${failed_command}' "

  if [ $status -eq 1 ]; then
    echo "failed"
  else
    echo "failed with status $status"
  fi
}

bats_frame_lineno() {
  local frame="$1"
  local lineno="${frame%% *}"
  echo "$lineno"
}

bats_frame_function() {
  local frame="$1"
  local rest="${frame#* }"
  local fn="${rest%% *}"
  echo "$fn"
}

bats_frame_filename() {
  local frame="$1"
  local rest="${frame#* }"
  local filename="${rest#* }"

  if [ "$filename" = "$BATS_TEST_SOURCE" ]; then
    echo "$BATS_TEST_FILENAME"
  else
    echo "$filename"
  fi
}

bats_extract_line() {
  local filename="$1"
  local lineno="$2"
  sed -n "${lineno}p" "$filename"
}

bats_strip_string() {
  local string="$1"
  printf "%s" "$string" | sed -e "s/^[ "$'\t'"]*//" -e "s/[ "$'\t'"]*$//"
}

bats_trim_filename() {
  local filename="$1"
  local length="${#BATS_CWD}"

  if [ "${filename:0:length+1}" = "${BATS_CWD}/" ]; then
    echo "${filename:length+1}"
  else
    echo "$filename"
  fi
}

bats_debug_trap() {
  if [ "$BASH_SOURCE" != "$1" ]; then
    bats_capture_stack_trace
  fi
}

bats_error_trap() {
  BATS_ERROR_STATUS="$?"
  BATS_ERROR_STACK_TRACE=( "${BATS_PREVIOUS_STACK_TRACE[@]}" )
  trap - debug
}

bats_teardown_trap() {
  trap "bats_exit_trap" exit
  local status=0
  teardown >>"$BATS_OUT" 2>&1 || status="$?"

  if [ $status -eq 0 ]; then
    BATS_TEARDOWN_COMPLETED=1
  elif [ -n "$BATS_TEST_COMPLETED" ]; then
    BATS_ERROR_STATUS="$status"
    BATS_ERROR_STACK_TRACE=( "${BATS_CURRENT_STACK_TRACE[@]}" )
  fi

  bats_exit_trap
}

bats_exit_trap() {
  local status
  local skipped
  trap - err exit

  skipped=""
  if [ -n "$BATS_TEST_SKIPPED" ]; then
    skipped=" # skip"
    if [ "1" != "$BATS_TEST_SKIPPED" ]; then
      skipped+=" ($BATS_TEST_SKIPPED)"
    fi
  fi

  if [ -z "$BATS_TEST_COMPLETED" ] || [ -z "$BATS_TEARDOWN_COMPLETED" ]; then
    echo "not ok $BATS_TEST_NUMBER $BATS_TEST_DESCRIPTION" >&3
    bats_print_stack_trace "${BATS_ERROR_STACK_TRACE[@]}" >&3
    bats_print_failed_command "${BATS_ERROR_STACK_TRACE[${#BATS_ERROR_STACK_TRACE[@]}-1]}" "$BATS_ERROR_STATUS" >&3
    sed -e "s/^/# /" < "$BATS_OUT" >&3
    status=1
  else
    echo "ok ${BATS_TEST_NUMBER}${skipped} ${BATS_TEST_DESCRIPTION}" >&3
    status=0
  fi

  rm -f "$BATS_OUT"
  exit "$status"
}

bats_perform_tests() {
  echo "1..$#"
  test_number=1
  status=0
  for test_name in "$@"; do
    "$0" $BATS_EXTENDED_SYNTAX "$BATS_TEST_FILENAME" "$test_name" "$test_number" || status=1
    let test_number+=1
  done
  exit "$status"
}

bats_perform_test() {
  BATS_TEST_NAME="$1"
  if [ "$(type -t "$BATS_TEST_NAME" || true)" = "function" ]; then
    BATS_TEST_NUMBER="$2"
    if [ -z "$BATS_TEST_NUMBER" ]; then
      echo "1..1"
      BATS_TEST_NUMBER="1"
    fi

    BATS_TEST_COMPLETED=""
    BATS_TEARDOWN_COMPLETED=""
    trap "bats_debug_trap \"\$BASH_SOURCE\"" debug
    trap "bats_error_trap" err
    trap "bats_teardown_trap" exit
    "$BATS_TEST_NAME" >>"$BATS_OUT" 2>&1
    BATS_TEST_COMPLETED=1

  else
    echo "bats: unknown test name '$BATS_TEST_NAME'" >&2
exit 1
fi
}

if [ -z "$TMPDIR" ]; then
BATS_TMPDIR="/tmp"
else
BATS_TMPDIR="${TMPDIR%/}"
fi

BATS_TMPNAME="$BATS_TMPDIR/bats.$$"
BATS_PARENT_TMPNAME="$BATS_TMPDIR/bats.$PPID"
BATS_OUT="${BATS_TMPNAME}.out"

bats_preprocess_source() {
BATS_TEST_SOURCE="${BATS_TMPNAME}.src"
{ tr -d '\r' < "$BATS_TEST_FILENAME"; echo; } | bats-preprocess > "$BATS_TEST_SOURCE"
trap "bats_cleanup_preprocessed_source" err exit
trap "bats_cleanup_preprocessed_source; exit 1" int
}

bats_cleanup_preprocessed_source() {
rm -f "$BATS_TEST_SOURCE"
}

bats_evaluate_preprocessed_source() {
if [ -z "$BATS_TEST_SOURCE" ]; then
BATS_TEST_SOURCE="${BATS_PARENT_TMPNAME}.src"
fi
source "$BATS_TEST_SOURCE"
}

exec 3<&1

if [ "$#" -eq 0 ]; then
bats_preprocess_source
bats_evaluate_preprocessed_source

if [ -n "$BATS_COUNT_ONLY" ]; then
echo "${#BATS_TEST_NAMES[@]}"
else
bats_perform_tests "${BATS_TEST_NAMES[@]}"
fi
else
bats_evaluate_preprocessed_source
bats_perform_test "$@"
fi`
	bats_format_tap_stream = `#!/usr/bin/env bash
set -e

# Just stream the TAP output (sans extended syntax) if tput is missing
command -v tput >/dev/null || exec grep -v "^begin "

header_pattern='[0-9]+\.\.[0-9]+'
IFS= read -r header

if [[ "$header" =~ $header_pattern ]]; then
  count="${header:3}"
  index=0
  failures=0
  skipped=0
  name=""
  count_column_width=$(( ${#count} * 2 + 2 ))
else
  # If the first line isn't a TAP plan, print it and pass the rest through
  printf "%s\n" "$header"
  exec cat
fi

update_screen_width() {
  screen_width="$(tput cols)"
  count_column_left=$(( $screen_width - $count_column_width ))
}

trap update_screen_width WINCH
update_screen_width

begin() {
  go_to_column 0
  printf_with_truncation $(( $count_column_left - 1 )) "   %s" "$name"
  clear_to_end_of_line
  go_to_column $count_column_left
  printf "%${#count}s/${count}" "$index"
  go_to_column 1
}

pass() {
  go_to_column 0
  printf " ✓ %s" "$name"
  advance
}

skip() {
  local reason="$1"
  [ -z "$reason" ] || reason=": $reason"
  go_to_column 0
  printf " - %s (skipped%s)" "$name" "$reason"
  advance
}

fail() {
  go_to_column 0
  set_color 1 bold
  printf " ✗ %s" "$name"
  advance
}

log() {
  set_color 1
  printf "   %s\n" "$1"
  clear_color
}

summary() {
  printf "\n%d test%s" "$count" "$(plural "$count")"

  printf ", %d failure%s" "$failures" "$(plural "$failures")"

  if [ "$skipped" -gt 0 ]; then
    printf ", %d skipped" "$skipped"
  fi

  printf "\n"
}

printf_with_truncation() {
  local width="$1"
  shift
  local string="$(printf "$@")"

  if [ "${#string}" -gt "$width" ]; then
    printf "%s..." "${string:0:$(( $width - 4 ))}"
  else
    printf "%s" "$string"
  fi
}

go_to_column() {
  local column="$1"
  printf "\x1B[%dG" $(( $column + 1 ))
}

clear_to_end_of_line() {
  printf "\x1B[K"
}

advance() {
  clear_to_end_of_line
  echo
  clear_color
}

set_color() {
  local color="$1"
  local weight="$2"
  printf "\x1B[%d;%dm" $(( 30 + $color )) "$( [ "$weight" = "bold" ] && echo 1 || echo 22 )"
}

clear_color() {
  printf "\x1B[0m"
}

plural() {
  [ "$1" -eq 1 ] || echo "s"
}

_buffer=""

buffer() {
  _buffer="${_buffer}$("$@")"
}

flush() {
  printf "%s" "$_buffer"
  _buffer=""
}

finish() {
  flush
  printf "\n"
}

trap finish EXIT

while IFS= read -r line; do
  case "$line" in
  "begin "* )
    let index+=1
    name="${line#* $index }"
    buffer begin
    flush
    ;;
  "ok "* )
    skip_expr="ok $index # skip (\(([^)]*)\))?"
    if [[ "$line" =~ $skip_expr ]]; then
      let skipped+=1
      buffer skip "${BASH_REMATCH[2]}"
    else
      buffer pass
    fi
    ;;
  "not ok "* )
    let failures+=1
    buffer fail
    ;;
  "# "* )
    buffer log "${line:2}"
    ;;
  esac
done

buffer summary`
	bats_preprocess = `#!/usr/bin/env bash
set -e

encode_name() {
  local name="$1"
  local result="test_"

  if [[ ! "$name" =~ [^[:alnum:]\ _-] ]]; then
    name="${name//_/-5f}"
    name="${name//-/-2d}"
    name="${name// /_}"
    result+="$name"
  else
    local length="${#name}"
    local char i

    for ((i=0; i<length; i++)); do
      char="${name:$i:1}"
      if [ "$char" = " " ]; then
        result+="_"
      elif [[ "$char" =~ [[:alnum:]] ]]; then
        result+="$char"
      else
        result+="$(printf -- "-%02x" \'"$char")"
      fi
    done
  fi

  echo "$result"
}

tests=()
index=0
pattern='^ *@test  *([^ ].*)  *\{ *(.*)$'

while IFS= read -r line; do
  let index+=1
  if [[ "$line" =~ $pattern ]]; then
    quoted_name="${BASH_REMATCH[1]}"
    body="${BASH_REMATCH[2]}"
    name="$(eval echo "$quoted_name")"
    encoded_name="$(encode_name "$name")"
    tests["${#tests[@]}"]="$encoded_name"
    echo "${encoded_name}() { bats_test_begin ${quoted_name} ${index}; ${body}"
  else
    printf "%s\n" "$line"
  fi
done

for test_name in "${tests[@]}"; do
  echo "bats_test_function ${test_name}"
done`
)

func WriteBats(path string) {
	ioutil.WriteFile(path + "/bats", []byte(bats), 0777)
	ioutil.WriteFile(path + "/bats-exec-suite", []byte(bats_exec_suite), 0777)
	ioutil.WriteFile(path + "/bats-exec-test", []byte(bats_exec_test), 0777)
	ioutil.WriteFile(path + "/bats-format-tap-stream", []byte(bats_format_tap_stream), 0777)
	ioutil.WriteFile(path + "/bats-preprocess", []byte(bats_preprocess), 0777)
}