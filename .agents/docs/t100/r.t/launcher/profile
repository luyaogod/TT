if [ "$1" = "" ]
   then
echo ""
echo "Please Select (1)正式區:topprd (2)測試區:toptst (3)標準區:topstd (4)Patch:toppth *)Exit [1]:"
echo ""
read ans

else
  ans="$1"
  case $ans in
  1) echo ""
     echo -e "(1)topprd:"""
     echo ""
  ;;
  2) echo ""
     echo -e "(2)toptst:"""
     echo ""
  ;;
  3) echo ""
     echo -e "(3)topstd:"""
     echo ""
  ;;
  4) echo ""
     echo -e "(4)toppth:"""
     echo ""
  ;;
  esac
fi
# for telnet only
EDITOR=vi;export EDITOR                         #SQLPlus Editor
stty erase ^H                                   #SQLPlus 倒退鍵
FGLSERVER=`who -m | cut -d'(' -f2 | cut -d')' -f1`;export FGLSERVER  #GDC Client IP

case $ans in
  1)
     ZONE=topprd;export ZONE
     ;;
  2)
     ZONE=toptst;export ZONE
     ;;
  3)
     ZONE=topstd;export ZONE
     ;;
  4)
     ZONE=toppth;export ZONE
     ;;
  *) exit
     ;;
esac
. /u1/etc/chenv
. /u1/etc/topenv

PATH=/u3/bin:$PATH
today=` date +"%a" `

readonly DGENV 2>/dev/null
readonly CUST 2>/dev/null
readonly TOPCHKOUT 2>/dev/null

echo "DGENV    = "$DGENV
echo "CUST     = "$CUST
echo "TOPENT   = "$TOPENT
echo "TOPALM   = "$TOPALM
echo "TOPCHKOUT= "$TOPCHKOUT
echo ""

