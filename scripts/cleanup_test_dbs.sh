#!/bin/bash

PREFIX="test_"

TEST_DB_LIST=$(psql "$DATABASE_URL" -l | awk '{ print $1 }' | grep '^[a-z]' | grep -v template | grep -v postgres)
for TEST_DB in $TEST_DB_LIST; do 
	if [ $(echo $TEST_DB | sed "s%^$PREFIX%%")  != $TEST_DB ] 
	then 
		echo "Dropping $TEST_DB"
		psql $DATABASE_URL -c "DROP DATABASE $TEST_DB"
		fi
	done
