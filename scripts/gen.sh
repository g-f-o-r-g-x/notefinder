#!/bin/sh

# I use this for tests when I need to populate a notebook
text=$(fortune)
name=$(echo "$text" | head -n 1)
echo "$text" > ~/Notes/$name
