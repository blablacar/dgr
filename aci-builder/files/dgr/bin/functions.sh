execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in "$fdir"/*; do
    [ -e "$file" ] && {
     	[ -x "$file" ] || chmod +x "$file"
		isLevelEnabled 4 && echo -e "\e[1m\e[32mRunning script -> $file\e[0m"
     	"$file"
    }
  done
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