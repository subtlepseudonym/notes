#!/bin/bash

add=0
del=0
IFS=$'\n'
for line in $(git diff-files --numstat); do
	add=$(($add + $(echo $line | cut -f1)))
	del=$(($del + $(echo $line | cut -f2)))
done

untracked=0
for line in $(git ls-files --others --exclude-standard); do
	untracked=$((untracked + $(wc -l $line | cut -f1 -d" ")))
done

version=$(git describe --abbrev=0)
short_rev=$(git rev-list -n1 --abbrev-commit HEAD)
rev_name=$(git name-rev --name-only HEAD)

echo "" | awk \
	-v version="${version}" \
	-v short_rev="${short_rev}" \
	-v rev_name="${rev_name}" \
	-v add="${add}" \
	-v del="${del}" \
	-v utd="${untracked}" \
'
function add_change(tag, count){
	if (!build_added) { v = v "+" }
	else if (!changes_added) { v = v "." }
	v = v tag count
	build_added=1
	changes_added=1
}

{
	v=version
	if (rev_name !~ /tags.*\^0/) {
		v = v "+" short_rev
		build_added=1
	}
	if (add != 0) { add_change("a", add) }
	if (del != 0) { add_change("d", del) }
	if (utd != 0) { add_change("u", utd) }
	printf v
}
'
