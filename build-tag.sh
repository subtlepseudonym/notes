#!/bin/bash

add=0
del=0
IFS=$'\n'
for line in $(git diff-files --numstat); do
	add=$(($add + $(echo $line | cut -f1)))
	del=$(($del + $(echo $line | cut -f2)))
done

version=$(git describe --abbrev=0)
short_rev=$(git rev-list -n1 --abbrev-commit HEAD)
branch=$(git name-rev --name-only HEAD)

echo "" | awk \
	-v version="${version}" \
	-v short_rev="${short_rev}" \
	-v branch="${branch}" \
	-v add="${add}" \
	-v del="${del}" \
'{
	fmt="%s"
	build_added=0
	if (branch != "master") {
		fmt = fmt "+" short_rev
		build_added=1
	}
	if (add != 0) {
		if (build_added) { fmt = fmt "." }
		else { fmt = fmt "+" }
		fmt = fmt "a" add
		build_added=1
	}
	if (del != 0) {
		if (build_added) { fmt = fmt "." }
		else { fmt = fmt "+" }
		fmt = fmt "d" del
		build_added=1
	}
	printf fmt, version, add, del
}'
