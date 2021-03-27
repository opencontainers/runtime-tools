command -v cat >/dev/null 2>/dev/null && test_set_prereq CAT
command -v echo >/dev/null 2>/dev/null && test_set_prereq ECHO
command -v head >/dev/null 2>/dev/null && test_set_prereq HEAD
command -v sed >/dev/null 2>/dev/null && test_set_prereq SED

true
