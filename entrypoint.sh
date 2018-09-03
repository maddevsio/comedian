#!/usr/bin/env bash

MYSQL_HOST=${MYSQL_HOST:-}
MYSQL_WAIT_TIMEOUT=${MYSQL_WAIT_TIMEOUT:- 60}
RUN_DB_MIGRATIONS=${RUN_DB_MIGRATIONS:-}

file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}

############################################################
###################CHECK IF MYSQL IS UP#####################
############################################################

if [[ ${MYSQL_HOST} ]]; then

  start_sec=$(date +%s)
  end_sec=$(date +%s)
  count=0

  while [ $((end_sec - start_sec)) -lt $MYSQL_WAIT_TIMEOUT ]; do
      (echo > /dev/tcp/$MYSQL_HOST/3306) >/dev/null 2>&1
      result=$?
      if [[ $result -eq 0 ]]; then
          let count+=1
          if [[ $count -ge 3 ]]; then
              >&2 echo "Mysql($MYSQL_HOST) is up - starting service"
              break
          fi
      else
          let count=0
      fi
      >&2 echo "Waiting if Mysql($MYSQL_HOST) is up.."
      sleep 5
      end_sec=$(date +%s)
  done

  if [[ $result -ne 0 ]]; then
      >&2 echo "Mysql($MYSQL_HOST) did not start after $MYSQL_WAIT_TIMEOUT sec"
      exit $result
  fi

fi

############################################################
###############RUN MIGRATION COMMAND IF SET#################
############################################################

if [[ ${RUN_DB_MIGRATIONS} == yes ]]; then
  echo "Running DATABASE migrations..."
  /comedian migrate
fi

# allow arguments to be passed to comedian daemon
if [[ ${1:0:1} = '-' ]]; then
  EXTRA_OPTS="$@"
  set --
fi

# default behaviour is to launch comedian
if [[ -z ${1} ]]; then
  echo "Starting diesel api demon..."
  exec /comedian ${EXTRA_OPTS}
else
  exec "$@"
fi
