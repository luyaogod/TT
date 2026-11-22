export FGLSQLDEBUG=9
export T100RUNWAIT=1
rm ${1}.log
r.r ${*} 2> ${1}.log
echo "${1}.log has been generated"
export FGLSQLDEBUG=0
export T100RUNWAIT=0
