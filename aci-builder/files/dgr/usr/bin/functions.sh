
echo_green() {
  echo -e "\033[0;32m${1}\033[0m"
}

echo_red() {
  echo -e "\033[0;31m${1}\033[0m"
}

echo_purple() {
  echo -e "\033[0;35m${1}\033[0m"
}

execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  if [ "$(ls -A "${fdir}")" ]; then
      for file in "${fdir}"/*; do
        if [ -f "$file" ]; then
            [ -e "$file" ] && {
                [ -x "$file" ] || chmod +x "$file"
                isLevelEnabled 4 && echo_green "Running script -> $file"
                "$file" || return 1
            }
        fi
      done
  fi
}

levelFromString() {
	case `echo ${1} | awk '{print toupper($0)}'` in
		"FATAL") echo 0; return 0 ;;
		"PANIC") echo 1; return 0 ;;
		"ERROR") echo 2; return 0 ;;
		"WARN"|"WARNING") echo 3; return 0 ;;
		"INFO") echo 4; return 0 ;;
		"DEBUG") echo 5; return 0 ;;
		"TRACE") echo 6; return 0 ;;
		*) echo 4 ;;
	esac
}

isLevelEnabled() {
	l=$(levelFromString $1)

	if [ $l -le $log_level ]; then
		return 0
	fi
	return 1
}

export log_level=$(levelFromString ${LOG_LEVEL:-INFO})