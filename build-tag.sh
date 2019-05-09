#!/bin/bash

add=0
del=0
IFS=$'\n'
for line in $(git diff-files --numstat); do
	add=$(($add + $(echo $line | cut -f1)))
	del=$(($del + $(echo $line | cut -f2)))
done
untracked=$(wc -l $(git ls-files --others --exclude-standard) | grep total | tr -s " " | cut -f2 -d" ")

version=$(git describe --abbrev=0)
short_rev=$(git rev-list -n1 --abbrev-commit HEAD)
rev_name=$(git name-rev --name-only HEAD)

echo "" | awk \
	-v version="${version}" \
	-v short_rev="${short_rev}" \
	-v rev_name="${rev_name}" \
	-v add="${add}" \
	-v del="${del}" \
	-v untracked="${untracked}" \
'{
	fmt="%s"
	build_added=0
	if (rev_name !~ /tags.*\^0/) {
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
	if (untracked != 0) {
		if (build_added) { fmt = fmt "." }
		else { fmt = fmt "+" }
		fmt = fmt "u" untracked
		build_added=1
	}
	printf fmt, version, add, del
}'
