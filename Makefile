watch:
	@inotifywait -qmr -e close_write . | \
		while read i; do clear; go build; done;
