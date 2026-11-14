######
# Run application whatever where you are
####

# Prompt how to work
if [ "$1" = "" ] ; then
   echo "Usage: r.r prog_name\n"
   exit 1
fi

PW=`pwd`
FGLRUN=${FGLRUN:=fglrun};export FGLRUN
cd $TEMPDIR
echo '\033[?7h\c'

i=1
cmd="$FGLRUN $ADZ/42r/"
while [ $i -le $# ]
do
 eval arg=\${$i}
 if [ $i = 1 ]; then
    cmd="${cmd}${arg} 'top-id'"
 else
    cmd="${cmd} '${arg}'"
 fi
 i=`expr $i + 1`
done
echo $cmd
eval $cmd
echo '\033[?7l\c'
cd $PW
