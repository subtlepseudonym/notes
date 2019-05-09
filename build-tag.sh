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

echo "" | awk -v version="${version}" -v add="${add}" -v del="${del}" '{
	fmt="%s"
	if (add != 0) { fmt = fmt ".a" add }
	if (del != 0) { fmt = fmt ".d" del }
	printf fmt, version, add, del
}'
