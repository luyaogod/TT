######
# Run application whatever where you are
####

PW=`pwd`
FGLRUN=${FGLRUN:=fglrun};export FGLRUN 
cd $TEMPDIR
#echo '\033[?7h\c'

#i=1
cmd="$FGLRUN $AZZ/42r/azzp000.42r"
#while [ $i -le $# ]
#do
# eval arg=\${$i}
# cmd="${cmd} '${arg}'"
# i=`expr $i + 1`
#done
#eval $cmd
$cmd "$@"
stus=$?

#echo '\033[?7l\c'
cd $PW

exit $stus
