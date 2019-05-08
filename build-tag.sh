#!/bin/bash

add=0
del=0
IFS=$'\n'
for line in $(git diff-files --numstat); do
	add=$(($add + $(echo $line | cut -f1)))
	del=$(($del + $(echo $line | cut -f2)))
done

git describe --abbrev=8 | tr -d g | cut -f1,3 -d- | tr - + | awk -v add="${add}" -v del="${del}" '{ printf("%s.a%s.d%s", $1, add, del) }'
