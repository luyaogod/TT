#該程式未解開Section, 採用最新樣板產出!
#該程式為freestyle程式!
{<section id="cl_xg.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0080(1900-01-01 00:00:00), PR版次:0080(2026-03-27 09:28:12)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000778
#+ Filename...: cl_xg
#+ Description: XG libary
#+ Creator....: 02716(2013-12-03 15:40:50)
#+ Modifier...: 00000 -SD/PR- 18060
 
{</section>}
 
{<section id="cl_xg.global" type="s" >}
#應用 p00 樣板自動產生(Version:6)
#add-point:填寫註解說明 name="main.memo"
#+ Modifier...: No.150508-00001#1 15/05/08 04010   讓MAC也能執行查詢報表
#+ Modifier...: No.150508-00001#3 15/05/26 04010   IE開不同視窗
#+ Modifier...: No.150611-00032#4 15/06/12 04010   表尾功能,排程mail address
#+ Modifier...: No.150709-00014#1 15/07/09 04010   1.串接訊息中心
#                                                  2.欄位權限控管改取角色設定的交集
#+ Modifier...: No.150803-00008#2 15/08/13 04010   1.報表匯出權限控管
#                                                  2.報表預設郵件清單
#+ Modifier...: No.150901-00021#1 15/09/01 04010   將adz中,使用到cl_gr跟cl_xg的FUNCTION獨立出來, 降低關聯性
#+ Modifier...: No.150904-00005#1 15/09/09 04010   列印條件中文化
#+ Modifier...: No.150904-00005#4 15/09/11 04010   gzgh046加入表尾說明
#+ Modifier...: No.150916-00002#1 15/09/16 02716   環境參數讀取資料庫
#+ Modifier...: No.150921-00060#5 15/10/01 04010   快速列印,單頭+明細可不設跳頁欄位
#+ Modifier...: No.150930-00022#2 15/10/01 04010   檢查欄位個數
#+ Modifier...: No.151027-00003#1 15/10/27 02716   判斷報表相關參數是否設定
#+ Modifier...: No.151027-00003#1 15/11/19 02716   cl_used_ap_hostname()刪除改用cl_get_ap_hostname
#+ Modifier...: No.151204-00012#1 15/12/04 04010   快速列印功能修正
#+ Modifier...: No.151216-00001#1 15/12/16 04010   排程問題修正
#+ Modifier...: No.160113-00004#2 16/01/13 04010   將郵件寄件人資訊寫入gzgh034
#+ Modifier...: No.160203-00001#1 16/02/03 04010   SQL修正,UNIQUE->DISTINCT
#+ Modifier...: No.160217-00006#1 16/02/17 04010   XG動態顯示報表名稱
#+ Modifier...: No.160330-00019#1 16/04/01 04010   1.多樣板過單需求
#                                                  2.g_success改g_rep_success
#+ Modifier...: No.160712-00018#1 16/07/12 04010   可切換樣板個數增加至10個
#+ Modifier...: No.160817-00023#2 16/08/24 04010   日期格式使用aooi100的設定,原本固化3種選項,需動態取得
#+ Modifier...: No.160921-00007#1 16/09/21 02716   報表名稱以報表作業(g_prog)，調整gzgh045欄位取得，及寄送mail部分
#+ Modifier...:                                    取消temptable varchar2欄位是null更新為空白,簡化處理時間
#+ Modifier...: No.160922-00012#2 16/10/11 04010   調整查詢報表取得樣板後，在IE上的排序方式，用gzgf002排序
#+ Modifier...: No.161103-00032#1 16/11/03 04010   因明細+單頭+子報表模板已合併至明細+單頭，故將gzgh010中template4的模板改指定template2
#+ Modifier...: No.161117-00037#1 16/11/17 04010   修正查詢報表製表日期,需參考時區
#                                                  資料庫空間不足時彈窗提示
#+ Modifier...: No.161222-00006#1 16/12/22 04010   將需要的最小項目移入 cl_s_aooi350，需搭配161222-00001#1
#+ Modifier...: No.161227-00007#1 16/12/27 02716   調整報表列印標題，報表title名稱等抓取報表元件，代號用報表作業
#+ Modifier...: No.170209-00022#1 17/02/09 04010   外部網路列印報表需求
#+ Modifier...: No.170210-00028#1 17/02/15 04010   指定log檔路徑
#+ Modifier...: No.170309-00023#1 17/03/09 04010   原外部IP列印報表需求, 不在4gl中判斷,可在T100登入首頁變更,故移除
#+ Modifier...: No.170301-00053#3 17/03/20 04010   1.查詢報表拋轉試算表模板
#                                                  2.取消檢查logo路徑FGLASIP,因已改將logo放在報表主機上
#+ Modifier...: No.170516-00054#1 17/05/16 04010   取得g_template的相關樣板
#+ Modifier...: No.170525-00105#1 17/05/25 04010   COUNT(*) 變更為 COUNT(1)
#+ Modifier...: No.170619-00003#1 17/06/19 09760   暫存表欄位型態與azzi300的設定是否相符
#+ Modifier...: No.170622-00040#1 17/06/23 09760   修正取得直接送印方式Bug(cl_xg_direct_print)
#+ Modifier...: No.170621-00048#2 17/07/13 04010   azzi850設定不可視欄位時，例如aapq380，分別有報表元件axrq380_x01與axrq380_x02
#                                                  兩個報表元件欄位編號都有l_docdt，無法單獨設置，故調整成axrq380_x01.l_docdt，用「.」分隔報表元件與欄位編號
#                                                  但資料庫現有欄位不會整批更新
#+ Modifier...: No.170828-00025#1 17/08/28 04010   1.補強排程錯誤訊息
#                                                  2.DISPLAY暫存表名稱時，加上"dsrept."，方便偵錯時直接撈取資料
#                                                  3.如果抓不到表頭表尾資訊則使用固化YY的內容，避免多語言資料未建立導致報表無法列印
#+ Modifier...: No.170911-00002#4 17/11/28 04010   報表郵件excel格式增加「只限資料」與「表頭表尾」選項(7、8、9、0)
#+ Modifier...: No.171205-00041#1 17/12/13 04010   修正動態指定欄位名稱(g_xg_fieldname),主報表的最後一筆欄位說明會影響到子報表所有說明
#+ Modifier...: No.180524-00013#1 18/05/29 11084   將預設報表元件帶入YY表頭表尾設定，新增判斷要帶入T100_YY或YY，用意為客戶家可能未更新資料庫，資料還停留在舊的
#+ Modifier...: No.180724-00009#2 18/07/31 11084   修改gzgf006的樣板名稱、控卡。
#+ Modifier...: No.180919-00051#1 18/09/19 04010   Windows XP的使用者，報表url改用chrome/firefox開啟。
#+ Modifier...: No.181017-00035#2 18/10/18 04010   優化CREATE TEMP TABLE，減少redo log用量，報表元件寫入資料時需搭配語法 INSERT /*+APPEND*/ INTO TABLE_NAME，否則設為nologging仍與一般table無異。
#+ Modifier...: No.180719-00037#1 18/11/02 11084   修正排程未依指定樣板寄送報表郵件、修改cl_xg_del_data FUNCTION DELETE TABLE段改用TRUNCATE確保資料表的資料會立即被刪除，避免佔用資源
#+ Modifier...: No.190107-00075#8 19/01/19 04010   郵件設定增加「進階設定」，提供郵件合併寄送功能
#+ Modifier...: No.190219-00005#1 19/03/12 04010   1.使用非IE瀏覽器替代方案(預設瀏覽器)
#                                                  2.log紀錄樣板名稱
#+ Modifier...: No.190426-00041#1 19/04/26 11084   修正客戶自行新增DB USER，無權限針對TABLE執行TRUNCATE，因此新增一個PROCEDURE只針對dsrept執行，避免權限過大問題
#+ Modifier...: No.190503-00009#2 19/05/03 11084   優化排程排除
#                                                  1.於XG LOG中新增START、END段落，以及LOG開始時間，以利後續排程加速後續排程問題排查
#                                                  2.將CALL cl_xg_view_logfile後，會RETURN的g_err_msg最前面加上"ERROR:"，排程工具撈取資料時才有依據
#+ Modifier...: No.190223-00010#2 19/05/14 11084   新增當寄送排程時，若無報表資料，則寄送無報表資料通知信件
#+ Modifier...: No.190408-00033#3 19/05/29 04010   1.T100V3.0判斷OS版本的視窗無法自動關閉，僅在T100V1.0判斷，最低需求為IE9
#                                                  2.因原先LIKE oojn_t的欄位型態未被SELECT使用，故將LIKE oojn_t型態的欄位調整為LIKE gzpe_t對應的欄位。
#+ Modifier...: No.190828-00035#1 19/08/28 04010   列印時預設使用XG，當資料量較大時，改用試算表模板。
#+ Modifier...: No.191127-00011#3 19/11/27 11084   調整日期格式呈現規則：azzi800>aooi100>azzi500
#+ Modifier...: No.200428-00006#1 20/04/28 12731   判斷相關權限
#+ Modifier...: No.230109-00023#1 23/01/31 18060   1.修正報表取得預設郵件清單的密件副本錯誤
#                                                  2.修正立即於背景執行的郵件主旨和內文無法寫入GZGH
#+ Modifier...: No.230825-00029#1 23/08/25 05292   MogDB相關lib調整(cl_abi、cl_abi_chk、cl_xg、cl_fr)
#+ Modifier...: No.231017-00022#1 23/10/17 05292   支持達夢版本timestamp與integer欄位類型
#+ Modifier...: No.231212-00003#1 23/12/12 18060   新增依據全域變數g_rep_filename 指定郵件寄送的報表附件名稱
#+ Modifier...: No.240130-00011#2 24/03/18 14910   1.修正Oracle跟Postgre表格大小寫，導致判斷欄位數異常問題
#                                          18060   2.修正數字/金額顯示格式異常問題
#                                          18060   3.郵件個資提示訊息改由aoos010系統參數判斷是否啟用
#+ Modifier...: No.240329-00013#1 24/03/29 18060   修正檢查欄位型態不一致異常問題
#+ Modifier...: No.240419-00012#1 24/04/19 18060   MogDB相關lib移除TABLESPACE使用自訂表空間(cl_xg、cl_xr)
#+ Modifier...: No.240612-00039#3 24/06/13 18060   支持KDB人大金倉版本調整alter table與numeric欄位類型
#+ Modifier...: No.260327-00001#1 26/03/27 18060   支持T100 AiGP架構運行XG報表


#end add-point
#add-point:填寫註解說明(客製用) name="main.memo_customerization"

#end add-point
#(ver:6) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:6) --- end ---
 
IMPORT os
#add-point:增加匯入項目 name="main.import"
IMPORT com
IMPORT xml
IMPORT util
#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
#add-point:增加匯入變數檔 name="global.inc"
GLOBALS "../../cfg/top_report.inc"        #T100報表的環境變數
GLOBALS "../../cfg/top_schedule.inc"      #T100排程的環境變數

GLOBALS
  DEFINE g_db_branch       STRING #POLARDB/MOGDB
END GLOBALS
#end add-point
 
{</section>}
 
{<section id="cl_xg.free_style_variable" type="s" >}
#add-point:free_style模組變數(Module Variable) name="free_style.variable"

#end add-point
 
{</section>}
 
{<section id="cl_xg.global_variable" type="s" >}
#add-point:自定義模組變數(Module Variable) name="global.variable"
GLOBALS 
   DEFINE gs_hostname   STRING               #AP SERVER Host Name
   DEFINE g_report      LIKE type_t.chr20
   DEFINE g_gzza003     LIKE gzza_t.gzza003  #歸屬模組
   DEFINE lc_sys_s      STRING               #原使用型態
   DEFINE channel       base.Channel
   DEFINE g_errlog      DYNAMIC ARRAY OF STRING
   DEFINE g_gzgh002	   DYNAMIC ARRAY OF LIKE gzgh_t.gzgh002   #組第一筆樣板ID的欄位說明  
END GLOBALS

type type_g_gzgf       RECORD 
   pick                VARCHAR,                   #顯示勾選
   gzgdl002            LIKE gzgdl_t.gzgdl002,     #樣板說明
   gzgf003             LIKE gzgf_t.gzgf003,       #客製否
   gzgf007             LIKE gzgf_t.gzgf002,       #樣板名稱        
   gzgf002             LIKE gzgf_t.gzgf002,       #樣板代號
   gzgf000             LIKE gzgf_t.gzgf000,       #報表樣板ID
   gzgf005             LIKE gzgf_t.gzgf005,       #用戶
   gzgf004             LIKE gzgf_t.gzgf004,       #角色   
   gzgf008             LIKE gzgf_t.gzgf008,       #預設樣板
   gzgf009             LIKE gzgf_t.gzgf009,       #列印簽核
   gzgf010             LIKE gzgf_t.gzgf010,       #列印簽核位置
   sort_num            LIKE type_t.num5           #排序碼(越大越優先)
   END RECORD

type type_g_gzgn       RECORD
   gzgn001             LIKE gzgn_t.gzgn001,       #使用者編號
   gzgn002             LIKE gzgn_t.gzgn002,       #職能角色編號
   gzgn003             LIKE gzgn_t.gzgn003,       #作業編號
   gzgn004             LIKE gzgn_t.gzgn004,       #報表樣板名稱
   gzgn005             LIKE gzgn_t.gzgn005,       #輸出種類
   gzgn006             LIKE gzgn_t.gzgn006,       #印表機名稱
   gzgn007             LIKE gzgn_t.gzgn007,       #列印份數
   gzgn008             LIKE gzgn_t.gzgn008,       #印表機參數
   gzgn009             LIKE gzgn_t.gzgn009,       #立即列印否
   gzgn010             LIKE gzgn_t.gzgn010,       #使用者端IP
   gzgf000             LIKE gzgf_t.gzgf000,       #報表樣板ID
   gzgf001             LIKE gzgf_t.gzgf001,       #報表元件代號
   gzgf002             LIKE gzgf_t.gzgf002,       #樣板代號
   gzgf003             LIKE gzgf_t.gzgf003,       #客製
   gzgf004             LIKE gzgf_t.gzgf004,       #角色
   gzgf005             LIKE gzgf_t.gzgf005,       #使用者
   gzgf008             LIKE gzgf_t.gzgf008,       #預設樣板
   gzgf009             LIKE gzgf_t.gzgf009,       #列印簽核
   gzgf010             LIKE gzgf_t.gzgf010,       #簽核位置
   gzgdl002            LIKE gzgdl_t.gzgdl002      #樣板說明
   END RECORD

TYPE type_g_attribute DYNAMIC ARRAY OF RECORD
                name         STRING,      #屬性名稱
                value        STRING       #屬性值
             END RECORD
             
DEFINE g_gzgf   DYNAMIC ARRAY OF  type_g_gzgf   

#直接送印變數
DEFINE g_dp_flag            STRING                #是否有設定直接送印
DEFINE g_dp                 type_g_gzgn
DEFINE g_rec_b              LIKE type_t.num10

DEFINE g_run_cmd      STRING,
       g_run_result   BOOLEAN,
       g_err_msg      STRING   

DEFINE g_date_fmt     STRING   #150904-00005#4 add
DEFINE g_url          STRING   #151216-00001#1 add
DEFINE g_ldate        LIKE type_t.dat   #161117-00037#1 add
DEFINE g_ip_type      STRING            #170209-00022#1 add

#170301-00053#3 add(s)
type type_g_gzim       RECORD
       pick           VARCHAR,                   #顯示勾選
       gziml004       LIKE gziml_t.gziml004,     #模板名稱
       gzim002        LIKE gzim_t.gzim002,       #模板編號
       gzim014        LIKE gzim_t.gzim014,       #客製
       gzim009_desc   LIKE oofa_t.oofa011,       #用戶名稱
       gzim010_desc   LIKE gzyal_t.gzyal003,     #角色名稱
       gzim000        LIKE gzim_t.gzim000,       #模板ID
       gzim009        LIKE gzim_t.gzim009,       #使用者
       gzim010        LIKE gzim_t.gzim010,       #角色
       gzim015        LIKE gzim_t.gzim015,       #預設樣板
       gzim007        LIKE gzim_t.gzim007,       #版本號
       sort_num       LIKE type_t.num5           #排序碼(越大越優先)       
   END RECORD
DEFINE g_gzim         DYNAMIC ARRAY OF  type_g_gzim
DEFINE g_rec_b1       LIKE type_t.num10
DEFINE g_ref_fields          DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
DEFINE g_rtn_fields          DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
#170301-00053#3 add(e)

DEFINE g_defBrowser   LIKE type_t.chr1           #190219-00005#1 add #預設瀏覽器

#end add-point
 
{</section>}
 
{<section id="cl_xg.other_dialog" type="s" >}

 
{</section>}
 
{<section id="cl_xg.other_function" readonly="Y" type="s" >}

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_xg_view ()
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_view()
   DEFINE l_url		     STRING
   DEFINE res		        LIKE type_t.num10
   DEFINE l_gzdel003      LIKE gzdel_t.gzdel003 #報表元件說明 #報表列印抬頭
   DEFINE l_time		     LIKE type_t.chr1000
   DEFINE l_gzgh		     RECORD LIKE gzgh_t.*  #XtraGrid中繼檔
   DEFINE l_gzgf		     RECORD LIKE gzgf_t.*  #XtraGrid報表樣板主檔
   DEFINE l_gzgf000	     LIKE gzgf_t.gzgf000   #報表樣板ID
   DEFINE l_gzgf006	     LIKE gzgf_t.gzgf006   #報表型態
   DEFINE l_gzgf007	     LIKE gzgf_t.gzgf007   #風格主題
   DEFINE l_gzgf013	     LIKE gzgf_t.gzgf013   #圖表軸數
   DEFINE l_gzgf014	     LIKE gzgf_t.gzgf014   #預設群組 
   DEFINE l_gzgf015	     LIKE gzgf_t.gzgf015   #預設排序  
   DEFINE l_gzgg020	     LIKE gzgg_t.gzgg020   #圖表類型
   DEFINE l_cnt,l_n,l_n2  LIKE type_t.num10
   DEFINE l_i             LIKE type_t.num10      
   DEFINE l_rep_cnt	     LIKE type_t.num5      #報表個數(主報表+子報表)
   #DEFINE l_gzgh002	     LIKE gzgh_t.gzgh002   #組第一筆樣板ID的欄位說明   #141016 mark  
   #DEFINE l_gzgf006_t     LIKE gzgf_t.gzgf006   #報表型態(組第一筆樣板ID的欄位說明) 
   #DEFINE l_gzgg020_t     LIKE gzgg_t.gzgg020   #圖表類型(組第一筆樣板ID的欄位說明)   #141016 mark
   #DEFINE l_gzgh012 	     LIKE gzgh_t.gzgh012   #自定義SQL
   DEFINE l_gzgh013 	     LIKE gzgh_t.gzgh013   #自定義连接
   DEFINE l_flag  	     LIKE type_t.chr1
   DEFINE l_certid        STRING
   DEFINE l_sql   	     STRING
   DEFINE l_sql2  	     STRING
   DEFINE l_str           STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token  	     STRING
   #DEFINE l_tables  	     DYNAMIC ARRAY OF STRING   
   DEFINE l_gzgg001       LIKE gzgg_t.gzgg001	#報表欄位代碼
   #DEFINE l_gzgg018       LIKE gzgg_t.gzgg018	#與主報表關聯鍵/自節點欄位
   DEFINE l_x             STRING
   DEFINE l_flds		     DYNAMIC ARRAY OF STRING
   DEFINE l_fld		     LIKE type_t.chr1000
   DEFINE l_fldstr		  STRING
   DEFINE l_logourl	     STRING
   DEFINE l_logopath	     STRING
   DEFINE l_cnt_type 	  LIKE type_t.num5      #計算圖表類型個數,判斷是否為複合圖形
   DEFINE l_ooef013       LIKE ooef_t.ooef013   #日期顯示格式
   DEFINE l_ooef023       LIKE ooef_t.ooef023   #數字/金額顯示格式    
   DEFINE l_tmps		     DYNAMIC ARRAY OF RECORD
                            gzgf000  LIKE gzgf_t.gzgf000, #樣板ID
                            gzgf006  LIKE gzgf_t.gzgf006, #報表型態
                            gzgf007  LIKE gzgf_t.gzgf007, #風格主題
                            #gzgf009  LIKE gzgf_t.gzgf009, #列印簽核   #141016 簽核 add #141022 mark                           
                            #gzgg020  LIKE gzgg_t.gzgg020, #圖表類型   #141016 mark
                            gzgf013  LIKE gzgf_t.gzgf013, #圖表軸數
                            gzgf002  LIKE gzgf_t.gzgf002, #樣板代號
                            gzgf004  LIKE gzgf_t.gzgf004, #用戶
                            gzgf005  LIKE gzgf_t.gzgf005  #角色
                          END RECORD
   DEFINE l_cust_flag     LIKE type_t.chr1      #客製否
   DEFINE l_cust_flag1    LIKE type_t.chr1      #客製否
   DEFINE l_success       LIKE type_t.num5      #成功與否
   DEFINE l_str1          STRING                
   DEFINE l_gzgh041       STRING
   DEFINE l_gzgj009       LIKE gzgj_t.gzgj009   #logo位置
   DEFINE l_ooek003       LIKE ooek_t.ooek003   #logo
   DEFINE l_oojj005_desc  LIKE type_t.chr500    #141016 add   
   DEFINE l_gzgm003       LIKE gzgm_t.gzgm003   #表頭代號
   DEFINE l_roles         STRING
   DEFINE l_gzgn007_str   STRING
   DEFINE l_seq_str       STRING
   DEFINE l_success_1     BOOLEAN
   #141230 add (s)
   DEFINE l_field_cnt     DYNAMIC ARRAY OF LIKE type_t.num10     #temp table欄位個數
   DEFINE l_gzgg_cnt      DYNAMIC ARRAY OF LIKE type_t.num10     #azzi300中的欄位個數
   DEFINE l_field_msg     STRING                                 #欄位個數錯誤訊息
   DEFINE l_msg           STRING
   DEFINE l_tmp_table     LIKE gzgh_t.gzgh003                    #暫存表名稱
   #141230 add (e)
   
   #150112 add(s)
   DEFINE ls_run_cmd      STRING,
          lb_run_result   BOOLEAN,
          ls_err_msg      STRING     
   #150112 add(e)  
   DEFINE li_idx          LIKE type_t.num5
   DEFINE l_ostype        STRING   #150508-00001#1 add
   DEFINE l_file_name     STRING   #150709-00014#1 add   
   DEFINE l_gzze003       LIKE gzze_t.gzze003   #150709-00014#1 add
   DEFINE l_pri_con       LIKE gzgj_t.gzgj009   #150904-00005#1 add #列印條件顯示否
   DEFINE l_cust_str      STRING                #160330-00019#1 add
   DEFINE l_sucess        BOOLEAN               #170301-00053#3 add
   DEFINE l_gzgf002_tmp   STRING                #170516-00054#1 add
   DEFINE l_version       STRING                #180919-00051#1 add #OS版本
   DEFINE l_gzpb009       LIKE gzpb_t.gzpb009   #180719-00037#1 add #紀錄樣板
   DEFINE ls_template     STRING                #190828-00035#1 add #模板代號   
   #260327-00001#1 add(s)
   DEFINE lc_str          STRING       
   DEFINE la_param        RECORD
                           prog      STRING,
                           param     DYNAMIC ARRAY OF STRING
                          END RECORD 
   #260327-00001#1 add(e)
   
   CALL g_errlog.clear()   #190219-00005#1 add   
   #170301-00053#3 add(s)   
   #試算表模板
   CALL cl_xg_chk_excel() RETURNING l_flag
   IF l_flag THEN
      CALL cl_xg_template_list(g_prog)
   
#      LET g_err_msg = "Execute Excel Template! "   #190503-00009#2 mark
#      LET g_err_msg = "ERROR:Execute Excel Template! "   #190503-00009#2 add   #190828-00035#1 mark
      LET g_err_msg = "INFO:Execute Excel Template! "   #190828-00035#1 add
      CALL cl_xg_view_addlog(g_err_msg)   #190219-00005#1 add
      CALL cl_xg_view_logfile()                    
      RETURN       
   END IF
   #170301-00053#3 add(e)
   
   #151027-00003#1 add -(s)
   #檢查環境變數
   IF cl_rpt_chk_env_exist() THEN
      LET INT_FLAG = TRUE
#      LET g_err_msg = "Please Check Env Exist! "   #151216-00001#1 add   #190503-00009#2 mark
      LET g_err_msg = "ERROR:Please Check Env Exist! "   #190503-00009#2 add
      CALL cl_xg_view_addlog(g_err_msg)   #190219-00005#1 add      
      CALL cl_xg_view_logfile()                    #151216-00001#1 add
      RETURN 
   END IF   
   #151027-00003#1 add -(e)  
    
   #170309-00023#1 mark(s) 
#   #170209-00022#1 add(s)
#   #判斷內網或外網
#   IF cl_xg_chk_internal_ip(FGL_GETENV("FGLSERVER")) THEN
#      LET g_ip_type = "XTRAGRIDIP"
#   ELSE
#      LET g_ip_type = "XTRAGRIDIP_E"   
#   END IF
#   #170209-00022#1 add(e)
   #170309-00023#1 mark(e)
   
   LET g_defBrowser = cl_get_para(g_enterprise,"","A-SYS-0081")   #190219-00005#1 add
   
   &ifdef v1   #190408-00033#3 add
   #180919-00051#1 add(s)
#   IF g_bgjob <> "Y" THEN   #190219-00005#1 mark
   IF g_bgjob <> "Y" AND g_defBrowser <> "Y" THEN   #190219-00005#1 add #使用預設瀏覽器就不檢查OS版本
      CALL cl_xg_osVersion() RETURNING l_version
   END IF
   #180919-00051#1 add(e)
   &endif   #190408-00033#3 add
   
   LET g_ldate = cl_get_current()                  #161117-00037#1 add

#   CALL g_errlog.clear()   #190219-00005#1 mark
   LET l_rep_cnt = g_rep_tmpname.getLength()
   LET l_roles = cl_replace_str(g_roles,",","','")
   LET l_roles = "('",l_roles,"')"   

   #直接送印
   LET g_dp_flag = "N"                       #151204-00012#1 add  #還原flag，避免先點快速列印再點列印時變數未清空   
   IF g_action_choice = "quickprint" THEN    #150921-00060#5 add  #點擊"快速列印"時才檢查直接送印
      CALL cl_xg_direct_print()
      IF INT_FLAG THEN RETURN END IF         #170622-00040#1 add
   END IF                                    #150921-00060#5 add 
   
   #子報表需要切換頁籤，因此只允許至多一個子報表
   IF l_rep_cnt = 2 THEN
      LET l_gzgh.gzgh028 = g_rep_tmpname[2]
   END IF

   #無自定義SQL，目前尚未完成此功能
   IF cl_null(g_xgrid.sql) THEN
      LET l_cnt = cl_xg_data_rowcnt(g_rep_tmpname[1])
      IF l_cnt = 0 THEN
#         LET g_err_msg = "There is no data! "         #151216-00001#1 add   #190503-00009#2 mark
         LET g_err_msg = "ERROR:There is no data! "   #190503-00009#2 add
         CALL cl_xg_view_addlog(g_err_msg)   #190219-00005#1 add         
         #170828-00025#1 add(s)
         IF g_bgjob = "Y" THEN
            DISPLAY "g_prog:",g_prog," g_rep_code:",g_rep_code," ",g_err_msg
         END IF
         #170828-00025#1 add(e)
         CALL cl_xg_view_logfile()                    #151216-00001#1 add               
         RETURN 
      END IF
      LET l_gzgh.gzgh003 = g_rep_tmpname[1]
   ELSE
      LET l_gzgh.gzgh003 = g_rep_tmpname[1]

      IF l_n2 > 0 THEN
         #LET l_gzgh.gzgh012 = l_x[l_n2+5,LENGTH(g_xgrid.sql)]
         #LET l_gzgh.gzgh012 = l_x(l_n2+5,LENGTH(l_x))  #WHERE Condition 
      ELSE
         LET l_gzgh.gzgh012 = ' 1=1'
      END IF

      #LET l_cnt = cl_xg_rowcnt(l_gzgh.gzgh003,l_gzgh.gzgh013=2) #檢查SQL筆數  #170622-00040#1 mark
      LET l_cnt = cl_xg_rowcnt(l_gzgh.gzgh003,l_gzgh.gzgh013) #檢查SQL筆數  #170622-00040#1 add
      IF l_cnt = 0 THEN 
#         LET g_err_msg = "There is no data(g_xgrid.sql)! "         #151216-00001#1 add   #190503-00009#2 mark
         LET g_err_msg = "ERROR:There is no data(g_xgrid.sql)! "   #190503-00009#2 add
         CALL cl_xg_view_addlog(g_err_msg)                         #190219-00005#1 add         
         CALL cl_xg_view_logfile()                                 #151216-00001#1 add                     
         RETURN 
      END IF
   END IF
  
   IF cl_null(gs_hostname) THEN
      #LET gs_hostname = cl_used_ap_hostname()  #151119-00011 mark
      LET gs_hostname = cl_get_ap_hostname()    #151119-00011 add
   END IF

   LET l_time = TIME
   LET l_gzgf000 = NULL
   IF cl_null(g_xgrid.rep_code) THEN LET g_xgrid.rep_code = g_rep_code END IF   

   #確認是否有客製樣板,有客製樣板只撈客製樣板,否則就只撈標準樣板
#   SELECT COUNT(*) INTO l_cnt FROM gzgf_t WHERE gzgfstus = "Y" AND gzgf001 = g_xgrid.rep_code AND (gzgf003 = "Y" OR gzgf003 = "c")   #170525-00105#1 mark
   SELECT COUNT(1) INTO l_cnt FROM gzgf_t WHERE gzgfstus = "Y" AND gzgf001 = g_xgrid.rep_code AND (gzgf003 = "Y" OR gzgf003 = "c")    #170525-00105#1 add
   IF l_cnt > 0 THEN
      LET l_cust_str = "'c'"         #160330-00019#1 add
#      LET l_cust_flag = "c"         #160330-00019#1 mark
      LET l_cust_flag1 = "Y"
   ELSE
      LET l_cust_str = "'s','x'"
#      LET l_cust_flag = "s"         #160330-00019#1 mark
      LET l_cust_flag1 = "N"      
   END IF

   #170516-00054#1 add(s)
   LET ls_template = g_template   #190828-00035#1 add
#   IF NOT cl_null(g_template) THEN   #190828-00035#1 mark
   IF NOT cl_null(g_template) AND ls_template.getIndexOf("_x",1) > 0 THEN   #190828-00035#1 add
      IF l_cust_str = "'c'" THEN
         LET l_gzgf002_tmp = g_template CLIPPED,"_c%"
      ELSE
         LET l_gzgf002_tmp = g_template CLIPPED,"_s%"
      END IF
   END IF
   #170516-00054#1 add(e)
   
   #順位1,屬於該用戶樣板
#   LET l_sql = "SELECT UNIQUE gzgf000,gzgf006,gzgf007,gzgf013,gzgf002,gzgf004,gzgf005,gzgf008 ",    #160203-00001#1 mark
   LET l_sql = "SELECT DISTINCT gzgf000,gzgf006,gzgf007,gzgf013,gzgf002,gzgf004,gzgf005,gzgf008 ",   #160203-00001#1 add
               "  FROM gzgf_t LEFT JOIN gzgg_t ON gzgg000 = gzgf000 AND gzgg002 = '",g_rlang,"'",
               " WHERE gzgfstus = 'Y' ",
               "   AND gzgf004 = '",g_user,"'",
#               "   AND (gzgf003 = '",l_cust_flag,"' OR gzgf003 = '",l_cust_flag1,"') "              #160330-00019#1 mark
               "   AND (gzgf003 IN (",l_cust_str,") OR gzgf003 = '",l_cust_flag1,"') "               #160330-00019#1 add
#   IF NOT cl_null(g_template) THEN
#      LET l_sql = l_sql," AND gzgf002 LIKE'",g_template,"%'"
#   END IF
   IF g_dp_flag = "Y" THEN 
#      IF NOT cl_null(g_template) THEN   #190828-00035#1 mark
      IF NOT cl_null(g_template) AND ls_template.getIndexOf("_x",1) > 0 THEN   #190828-00035#1 add
#         LET l_sql = l_sql," AND (gzgf002 LIKE'",g_template,"%' OR gzgf002 ='",g_dp.gzgn004 CLIPPED,"')"
         LET l_sql = l_sql," AND (gzgf002 = '",g_template,"' OR gzgf002 ='",g_dp.gzgn004 CLIPPED,"')"
      ELSE
         LET l_sql = l_sql," AND gzgf002 ='",g_dp.gzgn004 CLIPPED,"'"
      END IF      
   ELSE
#      IF NOT cl_null(g_template) THEN   #190828-00035#1 mark
      IF NOT cl_null(g_template) AND ls_template.getIndexOf("_x",1) > 0 THEN   #190828-00035#1 add
#         LET l_sql = l_sql," AND gzgf002 LIKE'",g_template,"%'"
#         LET l_sql = l_sql," AND gzgf002 = '",g_template,"'"                                                        #170516-00054#1 mark
         LET l_sql = l_sql," AND (gzgf002 = '",g_template CLIPPED,"' OR gzgf002 LIKE '",l_gzgf002_tmp CLIPPED,"') "   #170516-00054#1 add
      END IF   
   END IF

#   LET l_sql2 = " AND gzgf001 ='",g_xgrid.rep_code,"' ORDER BY gzgf008 DESC"           #160922-00012#2 mark
   LET l_sql2 = " AND gzgf001 ='",g_xgrid.rep_code,"' ORDER BY gzgf008 DESC,gzgf002 "   #160922-00012#2 add
   #拆解table name
   IF l_rep_cnt = 1 THEN
      LET l_sql2 = l_sql2 CLIPPED #," AND gzgf002 ='",g_rep_tmpname[1],"'" 
   ELSE 
      IF l_rep_cnt = 2  THEN #有子報表
         LET l_sql2 = l_sql2 CLIPPED #," AND gzgf002 IN ('",g_rep_tmpname[1],"'","'",g_rep_tmpname[2],"')"
      END IF  
   END IF
    
   LET l_sql = l_sql, l_sql2
   PREPARE gzgf_pre1 FROM l_sql
   DECLARE gzgf_cur1 CURSOR FOR gzgf_pre1

   #順位2,屬於該用戶所屬角色樣板
#   LET l_sql = "SELECT UNIQUE gzgf000,gzgf006,gzgf007,gzgf013,gzgf002,gzgf004,gzgf005, gzgf008 ",    #160203-00001#1 mark
   LET l_sql = "SELECT DISTINCT gzgf000,gzgf006,gzgf007,gzgf013,gzgf002,gzgf004,gzgf005, gzgf008 ",   #160203-00001#1 add
               "  FROM gzgf_t LEFT JOIN gzgg_t ON gzgg000 = gzgf000 AND gzgg002 = '",g_rlang,"'",
               " WHERE gzgfstus = 'Y' ",
               "   AND gzgf005 IN ",l_roles CLIPPED,
#               "   AND (gzgf003 = '",l_cust_flag,"' OR gzgf003 = '",l_cust_flag1,"') "              #160330-00019#1 mark
               "   AND (gzgf003 IN (",l_cust_str,") OR gzgf003 = '",l_cust_flag1,"') "               #160330-00019#1 add
#   IF NOT cl_null(g_template) THEN
#      LET l_sql = l_sql," AND gzgf002 LIKE'",g_template,"%'"
#   END IF               
   IF g_dp_flag = "Y" THEN 
#      IF NOT cl_null(g_template) THEN   #190828-00035#1 mark
      IF NOT cl_null(g_template) AND ls_template.getIndexOf("_x",1) > 0 THEN   #190828-00035#1 add
      
#         LET l_sql = l_sql," AND (gzgf002 LIKE'",g_template,"%' OR gzgf002 ='",g_dp.gzgn004 CLIPPED,"')"
         LET l_sql = l_sql," AND (gzgf002 = '",g_template,"' OR gzgf002 ='",g_dp.gzgn004 CLIPPED,"')"
      ELSE
         LET l_sql = l_sql," AND gzgf002 ='",g_dp.gzgn004 CLIPPED,"'"
      END IF      
   ELSE
#      IF NOT cl_null(g_template) THEN   #190828-00035#1 mark
      IF NOT cl_null(g_template) AND ls_template.getIndexOf("_x",1) > 0 THEN   #190828-00035#1 add
#         LET l_sql = l_sql," AND gzgf002 LIKE'",g_template,"%'"
#         LET l_sql = l_sql," AND gzgf002 = '",g_template,"'"                                                        #170516-00054#1 mark
         LET l_sql = l_sql," AND (gzgf002 = '",g_template CLIPPED,"' OR gzgf002 LIKE '",l_gzgf002_tmp CLIPPED,"') "   #170516-00054#1 add
      END IF   
   END IF
   #LET l_sql2 = " AND gzgf001 ='",g_xgrid.rep_code,"'"," ORDER BY gzgf008 DESC"
   LET l_sql2 = " AND gzgf001 ='",g_xgrid.rep_code,"'"," ORDER BY gzgf008 DESC,gzgf002,gzgf005 "   #160922-00012#2 add  

   #拆解table name
   IF l_rep_cnt = 1 THEN
      LET l_sql2 = l_sql2 CLIPPED #," AND gzgf002 ='",g_rep_tmpname[1],"'" 
   ELSE 
      IF l_rep_cnt = 2  THEN #有子報表
         LET l_sql2 = l_sql2 CLIPPED #," AND gzgf002 IN ('",g_rep_tmpname[1],"'","'",g_rep_tmpname[2],"')"
      END IF  
   END IF
  
   LET l_sql = l_sql, l_sql2
   PREPARE gzgf_pre2 FROM l_sql
   DECLARE gzgf_cur2 CURSOR FOR gzgf_pre2
    
   #順位3,default/default
#   LET l_sql = "SELECT UNIQUE gzgf000,gzgf006,gzgf007,gzgf013,gzgf002,gzgf004,gzgf005, gzgf008 ",    #160203-00001#1 mark
   LET l_sql = "SELECT DISTINCT gzgf000,gzgf006,gzgf007,gzgf013,gzgf002,gzgf004,gzgf005, gzgf008 ",   #160203-00001#1 add
               "  FROM gzgf_t LEFT JOIN gzgg_t ON gzgg000 = gzgf000 AND gzgg002 = '",g_rlang,"'",
               " WHERE gzgfstus = 'Y' ",
               "   AND gzgf004 = 'default'", 
               "   AND gzgf005 = 'default'",
#               "   AND (gzgf003 = '",l_cust_flag,"' OR gzgf003 = '",l_cust_flag1,"') "              #160330-00019#1 mark
               "   AND (gzgf003 IN (",l_cust_str,") OR gzgf003 = '",l_cust_flag1,"') "               #160330-00019#1 add
#   IF NOT cl_null(g_template) THEN
#      LET l_sql = l_sql," AND gzgf002 LIKE'",g_template,"%'"
#   END IF
   IF g_dp_flag = "Y" THEN 
#      IF NOT cl_null(g_template) THEN   #190828-00035#1 mark
      IF NOT cl_null(g_template) AND ls_template.getIndexOf("_x",1) > 0 THEN   #190828-00035#1 add
#         LET l_sql = l_sql," AND (gzgf002 LIKE'",g_template,"%' OR gzgf002 ='",g_dp.gzgn004 CLIPPED,"')"
         LET l_sql = l_sql," AND (gzgf002 = '",g_template,"' OR gzgf002 ='",g_dp.gzgn004 CLIPPED,"')"
      ELSE
         LET l_sql = l_sql," AND gzgf002 ='",g_dp.gzgn004 CLIPPED,"'"
      END IF      
   ELSE
#      IF NOT cl_null(g_template) THEN   #190828-00035#1 mark
      IF NOT cl_null(g_template) AND ls_template.getIndexOf("_x",1) > 0 THEN   #190828-00035#1 add      
#         LET l_sql = l_sql," AND gzgf002 LIKE'",g_template,"%'"
#         LET l_sql = l_sql," AND gzgf002 = '",g_template,"'"                                                        #170516-00054#1 mark
         LET l_sql = l_sql," AND (gzgf002 = '",g_template CLIPPED,"' OR gzgf002 LIKE '",l_gzgf002_tmp CLIPPED,"') "   #170516-00054#1 add
      END IF   
   END IF
   
   #160922-00012#2 add(s)
   LET l_sql2 = " AND gzgf001 ='",g_xgrid.rep_code,"'"," ORDER BY gzgf008 DESC,gzgf002 "

   #拆解table name
   IF l_rep_cnt = 1 THEN
      LET l_sql2 = l_sql2 CLIPPED #," AND gzgf002 ='",g_rep_tmpname[1],"'" 
   ELSE 
      IF l_rep_cnt = 2  THEN #有子報表
         LET l_sql2 = l_sql2 CLIPPED #," AND gzgf002 IN ('",g_rep_tmpname[1],"'","'",g_rep_tmpname[2],"')"
      END IF  
   END IF
   #160922-00012#2 add(e)
   
   LET l_sql = l_sql,l_sql2
   PREPARE gzgf_pre3 FROM l_sql
   DECLARE gzgf_cur3 CURSOR FOR gzgf_pre3

   LET l_cnt = 1
   
   #先取此 用戶 的樣板(g_user/default)
   FOREACH gzgf_cur1 INTO l_tmps[l_cnt].*
      IF STATUS THEN INITIALIZE g_errparam TO NULL
         LET g_errparam.code = STATUS
         LET g_errparam.extend = 'gzgf_cur1'
         LET g_errparam.popup = TRUE
         CALL cl_err()
      END IF
         LET l_cnt = l_cnt + 1
   END FOREACH    
    
   #再取此 角色 的樣板(default/g_site)
   FOREACH gzgf_cur2 INTO l_tmps[l_cnt].*
      IF STATUS THEN INITIALIZE g_errparam TO NULL
         LET g_errparam.code = STATUS
         LET g_errparam.extend = 'gzgf_cur2'
         LET g_errparam.popup = TRUE
         CALL cl_err()
      END IF
         LET l_cnt = l_cnt + 1
   END FOREACH
    
   #再取標準的樣板(default/default)
   FOREACH gzgf_cur3 INTO l_tmps[l_cnt].*
      IF STATUS THEN INITIALIZE g_errparam TO NULL
         LET g_errparam.code = STATUS
         LET g_errparam.extend = 'gzgf_cur3'
         LET g_errparam.popup = TRUE
         CALL cl_err()
      END IF
         LET l_cnt = l_cnt + 1
   END FOREACH 

   #CALL l_tmps.deleteElement(l_cnt)
   #180719-00037#1 add(s)
   IF g_bgjob = "Y" THEN
      SELECT gzpb009 INTO l_gzpb009 FROM gzpb_t WHERE gzpbent = g_enterprise AND gzpb001 = g_schedule_id AND gzpb002 = g_schedule_seq
      IF l_gzpb009 IS NOT NULL THEN
         LET g_reportid = l_gzpb009
      END IF
      IF NOT cl_null(g_reportid) THEN
         CALL l_tmps.clear()
         SELECT gzgf000,gzgf006,gzgf007,gzgf013,gzgf002,gzgf004,gzgf005 INTO l_tmps[1].* FROM gzgf_t WHERE gzgf000 = g_reportid 
      END IF
   END IF
   #180719-00037#1 add(e)
   #模板名稱
   #LET l_gzgh002= l_tmps[1].gzgf000   #141016 mark
   #LET l_gzgg020_t = l_tmps[1].gzgg020   #141016 mark

   SELECT * INTO l_gzgf.* FROM gzgf_t WHERE gzgf000 = l_tmps[1].gzgf000 AND gzgfstus = 'Y'
   #141230 add (s)
   LET l_n = 1
   LET l_field_msg = NULL
   
   FOR l_n = 1 TO g_rep_tmpname.getlength()
      #取得temp table中的欄位個數
      LET l_tmp_table = g_rep_tmpname[l_n]
      SELECT COUNT(COLUMN_NAME) INTO l_field_cnt[l_n] FROM all_tab_columns
       #WHERE lower(table_name) = l_tmp_table AND lower(owner) = g_report   #240130-00011#2 mark
       WHERE table_name = upper(l_tmp_table) AND owner = upper(g_report)   #240130-00011#2 add
#      DISPLAY "TEMP TABLE欄位個數:",l_field_cnt[l_n]

      CASE l_n
         WHEN 1
#            SELECT COUNT(gzgg001) INTO l_gzgg_cnt[l_n] FROM gzgg_t WHERE gzgg000 = l_tmps[1].gzgf000 AND gzgg002 = g_rlang AND (gzgg017 IS NULL OR gzgg017 = 'Y')   #150930-00022#2 mark   
            SELECT COUNT(gzgg001) INTO l_gzgg_cnt[l_n] FROM gzgg_t WHERE gzgg000 = l_tmps[1].gzgf000 AND gzgg002 = g_rlang AND (gzgg017 IS NULL OR gzgg017 = 'Y' OR gzgg017 = 'N')   #150930-00022#2 add
         WHEN 2
            SELECT COUNT(gzgg001) INTO l_gzgg_cnt[l_n] FROM gzgg_t WHERE gzgg000 = l_tmps[1].gzgf000 AND gzgg002 = g_rlang AND gzgg017 = '2'
      END CASE
     
#      DISPLAY "azzi300欄位個數:",l_gzgg_cnt[l_n]

      IF l_field_cnt[l_n] > l_gzgg_cnt[l_n] THEN
         CALL cl_xg_chk_tmp_table_field("g",l_n,l_tmps[1].gzgf000) RETURNING l_msg
         LET l_msg = cl_getmsg("lib-00239",g_lang) CLIPPED," > ",cl_getmsg("lib-00240",g_lang) CLIPPED,"\r\n",l_msg
         LET l_field_msg = l_field_msg,"\r\n",l_msg       
      END IF
      IF l_field_cnt[l_n] < l_gzgg_cnt[l_n] THEN
         CALL cl_xg_chk_tmp_table_field("l",l_n,l_tmps[1].gzgf000) RETURNING l_msg
         LET l_msg = cl_getmsg("lib-00239",g_lang) CLIPPED," < ",cl_getmsg("lib-00240",g_lang) CLIPPED,"\r\n",l_msg
         LET l_field_msg = l_field_msg,"\r\n",l_msg       
      END IF      
      #170619-00003#1 add(s)
      IF l_field_cnt[l_n] = l_gzgg_cnt[l_n] THEN
         CALL cl_xg_chk_tmp_table_field_type(l_n,l_tmps[1].gzgf000) RETURNING l_msg
         IF NOT cl_null(l_msg) THEN
            LET l_msg = cl_getmsg("lib-00434",g_lang) CLIPPED,"\r\n",l_msg
            LET l_field_msg = l_field_msg,"\r\n",l_msg
         END IF
      END IF
      #170619-00003#1 end(e)
   END FOR  
   
   IF NOT cl_null(l_field_msg) THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'lib-00243'
      LET g_errparam.extend = l_field_msg
      LET g_errparam.popup = TRUE
      CALL cl_err()

      LET INT_FLAG = TRUE
      LET g_err_msg = l_field_msg            #151216-00001#1 add
      #190503-00009#2 mark(s)
      #190219-00005#1 add(s)
#      LET g_err_msg = g_err_msg,"\nTemplate gzgf000:",l_tmps[1].gzgf000,",gzgf002:",l_tmps[1].gzgf002,",gzgf004:",l_tmps[1].gzgf004,
#                      ",gzgf005:",l_tmps[1].gzgf005
      #190503-00009#2 mark(e)
      #190503-00009#2 add(s)
      LET g_err_msg = "ERROR:",g_err_msg,"\nTemplate gzgf000:",l_tmps[1].gzgf000,",gzgf002:",l_tmps[1].gzgf002,
                      ",gzgf004:",l_tmps[1].gzgf004,",gzgf005:",l_tmps[1].gzgf005
      #190503-00009#2 add(e)
      CALL cl_xg_view_addlog(g_err_msg)
      #190219-00005#1 add(e)
      CALL cl_xg_view_logfile()              #151216-00001#1 add
      RETURN      
   END IF    
   
#   SELECT COUNT(*) INTO l_field_cnt[1] FROM all_tab_columns 
#    WHERE lower(table_name) = l_gzgh.gzgh003 AND lower(owner) = g_report
#    WHERE lower(table_name) = l_gzgh.gzgh003 AND lower(owner) = g_report     
#   SELECT COUNT(*) INTO l_gzgg_cnt[1] FROM gzgg_t WHERE gzgg000 = l_tmps[1].gzgf000 AND gzgg002 = g_rlang   
   
   #141230 add (e)   
#   #191127-00011#3 mark(s)
#   #從組織基本資料維護作業(aooi100)中取出設定的日期顯示格式(l_ooef013)和數字/金額顯示格式(l_ooef023)
#   SELECT ooef013, ooef023 INTO l_ooef013, l_ooef023
#     FROM ooef_t
#    WHERE ooefstus = "Y" AND ooefent = g_enterprise AND ooef001 = g_site 
#
#   #160817-00023#2 add(s)
#   #日期格式的SCC碼是10
#   SELECT gzcbl004 INTO l_gzgh.gzgh038 FROM gzcbl_t WHERE gzcbl001 = 10 AND gzcbl002 = l_ooef013 AND gzcbl003 = g_lang
#   LET l_str = l_gzgh.gzgh038
   #191127-00011#3 mark(e)
   CALL cl_rpt_get_date_fmt() RETURNING l_str   #191127-00011#3 add
   #如果大小寫不一致,該欄位會變0,故需轉換
   LET l_str = cl_str_replace(l_str, "Y", "y")
   LET l_str = cl_str_replace(l_str, "m", "M")
   LET l_str = cl_str_replace(l_str, "D", "d")
   IF cl_null(l_str) THEN LET l_str = "yyyy/MM/dd" END IF
   LET l_gzgh.gzgh038 = l_str
   #160817-00023#2 add(e)
   #160817-00023#2 mark(s)
#   CASE l_ooef013
#      WHEN "1"
#         LET l_gzgh.gzgh038 = "yyyy/MM/dd"
#      WHEN "2"
#         LET l_gzgh.gzgh038 = "MM/dd/yyyy"
#      WHEN "3"
#         LET l_gzgh.gzgh038 = "dd/MM/yyyy"
#      OTHERWISE
#         LET l_gzgh.gzgh038 = "yyyy/MM/dd"  
#   END CASE
   #160817-00023#2 mark(e)
   LET g_date_fmt = l_gzgh.gzgh038   #150904-00005#4 add
   
   #240130-00011#2 add(s)
   SELECT ooef023 INTO l_ooef023
     FROM ooef_t
     WHERE ooefstus = "Y" AND ooefent = g_enterprise AND ooef001 = g_site 
   #240130-00011#2 add(e)

   #數字/金額顯示格式,gzgh039和gzgh040都不可為空值
   CASE l_ooef023
      WHEN "1"
         LET l_gzgh.gzgh039 = "."  
         LET l_gzgh.gzgh040 = ","  
      WHEN "2"
         LET l_gzgh.gzgh039 = ","   
         LET l_gzgh.gzgh040 = "."  
      OTHERWISE    
         LET l_gzgh.gzgh039 = "."  
         LET l_gzgh.gzgh040 = ","  
   END CASE

   LOCATE l_gzgh.gzgh041 IN FILE
   LOCATE l_gzgh.gzgh046 IN FILE                                  #表頭表尾資訊
   CALL cl_xg_get_gzgh046() RETURNING l_gzgh.gzgh046 ,l_gzgm003   #表頭表尾資訊
   CALL cl_xg_get_gzgh047() RETURNING l_gzgh.gzgh047              #不可視欄位
   
#   #開啟IE後，至多提供三個樣板做切換   #160712-00018#1 mark 
#   FOR l_n = 1 TO 3                 #160712-00018#1 mark 
   #開啟IE後，至多提供十個樣板做切換    #160712-00018#1 add
   FOR l_n = 1 TO 10                 #160712-00018#1 add
      IF cl_null(l_tmps[l_n].gzgf000) THEN CONTINUE FOR END IF
      LET g_gzgh002[l_n] = l_tmps[l_n].gzgf000
      #CALL cl_xg_chk_column(l_tmps[l_n].gzgf000,g_xgrid.sql,l_gzgh.gzgh003,l_gzgh.gzgh012)  #170622-00040#1 mark
#      IF g_success = 'N' THEN EXIT FOR END IF      #160330-00019#1 mark
      IF g_rep_success = 'N' THEN EXIT FOR END IF   #160330-00019#1 add
      
      IF l_flag = 'N' THEN
         SELECT gzgf014,gzgf015 INTO l_gzgf014,l_gzgf015  #群組,排序
           FROM gzgf_t
          WHERE gzgf000 = l_tmps[l_n].gzgf000
            AND gzgfstus = 'Y'

         LET l_flag = 'Y'
         #取得逾時設定，以首筆樣板的設定為主
         SELECT gzgf011 INTO l_gzgf.gzgf011 FROM gzgf_t
          WHERE gzgf000 = l_gzgf000
            AND gzgfstus = 'Y' 
      END IF

      LET l_gzgf000 = l_tmps[l_n].gzgf000
      LET l_gzgf006 = l_tmps[l_n].gzgf006
      LET l_gzgf007 = l_tmps[l_n].gzgf007
#      SELECT gzgg020 INTO l_gzgg020 FROM gzgg_t WHERE gzgg000 = l_gzgf000   #141016 add
      #180724-00009#2 add(s)
      IF l_gzgf006 NOT MATCHES "[1234567]" THEN
         LET g_errparam.extend = ""
         LET g_errparam.code = "lib-00640"
         LET g_errparam.popup = "TRUE"
         CALL cl_err()
         RETURN
      END IF
      #180724-00009#2 add(e)
      #LET l_gzgg020 = l_tmps[l_n].gzgg020   #141016 mark  
      LET l_gzgf013 = l_tmps[l_n].gzgf013
      
      LET l_gzgh.gzgh002 = l_gzgh.gzgh002 CLIPPED,"|",l_tmps[l_n].gzgf000 #可用的樣板ID
      LET l_gzgh.gzgh014 = l_gzgh.gzgh014 CLIPPED,l_tmps[l_n].gzgf007 USING '<<<<','|' #可用的風格

      CALL cl_xg_get_gzgh011(l_tmps[l_n].gzgf006,l_tmps[l_n].gzgf000,l_tmps[l_n].gzgf004,l_tmps[l_n].gzgf005,l_tmps[l_n].gzgf002,l_gzgg020) RETURNING l_str1
      LET l_gzgh.gzgh011 = l_gzgh.gzgh011 CLIPPED,"|",l_str1 #可用的樣板型態中文說明
      IF NOT cl_null(g_schedule_id) THEN
         LET l_seq_str = g_schedule_seq
         LET l_gzgh.gzgh048 = g_schedule_id,"|",l_seq_str.trim()   #排程資訊
      END IF

      LET l_str1 = cl_xg_get_gzgh041(l_tmps[l_n].gzgf000)
      LET l_gzgh041 = l_gzgh041,"{+}",l_str1
      
      IF cl_null(g_xgrid.sql) THEN      
###141016 mark (s)      
#         IF NOT cl_null(l_gzgg020) THEN
            #報表型態
            #IF cl_null(l_gzgf006) THEN 
            #   DISPLAY "請檢査gzgf006是否正確"
            #END IF
#            CASE l_gzgf006 
#               WHEN "1" LET l_gzgh.gzgh010 = l_gzgh.gzgh010 CLIPPED,'template11.aspx','|'
#               WHEN "2" LET l_gzgh.gzgh010 = l_gzgh.gzgh010 CLIPPED,'template12.aspx','|'
#               OTHERWISE 
#                  DISPLAY "請檢査gzgf006是否正確"                 
#                  INITIALIZE g_errparam TO NULL
#                  LET g_errparam.code = 'azz-00149'
#                  LET g_errparam.extend = ''
#                  LET g_errparam.popup = TRUE
#                  CALL cl_err()

#                  CALL cl_xg_view_addlog(cl_getmsg('azz-00149',g_rlang))
#                  LET g_success = 'N' 
#                  EXIT FOR
#            END CASE
###141016 mark (e)
#180724-00009#2 mark(s)
####141016 add (s)
#         IF l_gzgf006 = "7" THEN   #明細清單+圖形
#            LET l_gzgh.gzgh010 = l_gzgh.gzgh010 CLIPPED,'template11.aspx','|'
####141016 add (e)
#180724-00009#2 mark(e)
         #180724-00009#2 add(s)
         IF l_gzgf006 = "7" THEN   #明細清單(多行單身)
            LET l_gzgh.gzgh010 = l_gzgh.gzgh010 CLIPPED,'template7.aspx','|'
         #180724-00009#2 add(e)
         ELSE
            #161103-00032#1 add(s)
            IF l_gzgf006 = "4" THEN
               LET l_gzgf006 = "2"
            END IF   
            #161103-00032#1 add(e)
            LET l_gzgh.gzgh010 = l_gzgh.gzgh010 CLIPPED,'template',l_gzgf006 USING '#','.aspx','|'
         END IF
      ELSE
         CASE l_gzgf006
            WHEN "1" LET l_gzgh.gzgh010 = l_gzgh.gzgh010 CLIPPED,'template21.aspx','|'
            WHEN "2" LET l_gzgh.gzgh010 = l_gzgh.gzgh010 CLIPPED,'template22.aspx','|'
         END CASE
      END IF #cl_null(g_xgrid.sql)
   END FOR
   LET l_gzgh041 = l_gzgh041.substring(4,LENGTH(l_gzgh041))
#   LET l_gzgh.gzgh041 = l_gzgh041  #150402 mark  
   LET l_gzgh.gzgh041 = NULL #150402 add
   
   #IF g_success = 'N' THEN RETURN END IF
   IF cl_null(l_gzgf000) THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'azz-00145'
      LET g_errparam.extend = ''
      LET g_errparam.popup = TRUE
      CALL cl_err()
 
      #CALL cl_err('報表設定作業未設定,請檢查','',1)
      #CALL cl_xg_view_addlog(cl_getmsg('azz1270',g_lang))
      #RETURN
#      LET g_success = 'N'      #160330-00019#1 mark
      LET g_rep_success = 'N'   #160330-00019#1 add
   END IF
   
#   IF g_success = 'N' THEN       #160330-00019#1 mark
   IF g_rep_success = 'N' THEN    #160330-00019#1 add
      CALL cl_xg_view_logfile()
      RETURN 
   END IF

   CALL cl_xg_set_url() RETURNING l_certid,l_url,l_gzgh013
   LET g_url = l_url   #151216-00001#1 add
   #DISPLAY "l_url: ", l_url  
   
   LET l_gzgh.gzgh002 = l_gzgh.gzgh002[2,LENGTH(l_gzgh.gzgh002)] 
   LET l_gzgh.gzgh010 = l_gzgh.gzgh010[1,LENGTH(l_gzgh.gzgh010)-1]
   LET l_gzgh.gzgh014 = l_gzgh.gzgh014[1,LENGTH(l_gzgh.gzgh014)-1]  
   LET l_gzgh.gzgh011 = l_gzgh.gzgh011[2,LENGTH(l_gzgh.gzgh011)]  
   IF g_dp_flag = 'Y' THEN
      LET l_gzgn007_str = g_dp.gzgn007
      LET l_gzgn007_str = l_gzgn007_str.trim()
      LET l_gzgh.gzgh036 = g_dp.gzgn005 CLIPPED,"|",g_dp.gzgn006 CLIPPED,"|",l_gzgn007_str,"|",g_dp.gzgn008 CLIPPED,"|",g_dp.gzgn009 CLIPPED
   END IF
   
   #160113-00004#2 add(s)
   #取得寄件者資訊
   CALL cl_gr_get_mailsender()
   LET l_gzgh.gzgh034 = g_xml.sender 
   #160113-00004#2 add(e)
   
   #取得匯出檔名
   #CALL cl_xg_view_export_filename(l_gzgf000) RETURNING l_gzgh.gzgh035 #231212-00003#1 mark
   #231212-00003#1 add(s)
   IF cl_null(g_rep_filename) THEN
      CALL cl_xg_view_export_filename(l_gzgf000) RETURNING l_gzgh.gzgh035
   ELSE
      LET l_gzgh.gzgh035 = g_rep_filename
   END IF
   #231212-00003#1 add(e)
   

   #檢查跳頁 欄位合理性 --start
   #邏輯: 單頭單身模板僅允許1個跳頁欄位，若超過1個時，則4gl需改為複合欄位
   #例: 資料依 年+月 跳頁，則4gl建立TEMP TABLE時必須建立一個 "年+月"的欄位
   IF l_gzgh.gzgh010 = 'template2.aspx' THEN
#      SELECT COUNT(*) INTO l_cnt FROM gzgg_t   #170525-00105#1 mark
      SELECT COUNT(1) INTO l_cnt FROM gzgg_t    #170525-00105#1 add
       WHERE gzgg000 = l_gzgf000
         #AND gzgg001 = g_xgrid.skippage_field
         AND gzgg001 = l_gzgf.gzgf016
      #150921-00060#5 mark(s)   
      #單頭+明細不一定要有跳頁欄位
#      IF l_cnt = 0 THEN
#         INITIALIZE g_errparam TO NULL
#         LET g_errparam.code = 'azz-00147'
#         LET g_errparam.extend = l_gzgf.gzgf016
#         LET g_errparam.popup = TRUE
#         CALL cl_err()
#
#         CALL cl_xg_view_addlog(cl_getmsg('azz-00147',g_rlang))
#         RETURN
#      END IF
      #150921-00060#5 mark(e)   
   ELSE
      IF NOT cl_null(l_gzgf.gzgf016) THEN
         #IF g_xgrid.skippage_field <> g_xgrid.grup_field THEN
         #   LET g_xgrid.skippage_field = g_xgrid.grup_field 
         #END IF
         #IF NOT cl_null(g_xgrid.grup_field) THEN
         IF NOT cl_null(l_gzgf.gzgf014) THEN 
            IF l_gzgf.gzgf016 <> l_gzgf.gzgf014 THEN  
               LET l_gzgf.gzgf016 = l_gzgf.gzgf014  
            END IF
         ELSE
            LET l_gzgf.gzgf014 = l_gzgf.gzgf016
         END IF
      END IF
   END IF
   #檢查跳頁 欄位合理性 --end

   #檢查GROUP 欄位合理性 --start
   #調整抓取邏輯，優先順序 : 1.習慣設置(gzgf014) 2.預設群組(gzgf014)
   IF cl_null(l_gzgf.gzgf014) THEN
      CALL cl_xg_view_gzgf014(l_gzgf000) RETURNING l_gzgf.gzgf014 #無習慣設定，改取gzgf014
   END IF

   LET l_gzgf.gzgf014 = DOWNSHIFT(l_gzgf.gzgf014) 
   LET lst_token = base.StringTokenizer.create(l_gzgf.gzgf014, ",")
   LET l_n = 1
   CALL l_flds.clear()
   WHILE lst_token.hasMoreTokens()
      LET l_flds[l_n] = lst_token.nextToken()
      #170622-00040#1 mark(s)
#      FOR l_cnt = 1 TO l_flds.getlength()
#         IF l_cnt = l_n THEN CONTINUE FOR END IF
#         IF l_flds[l_n]=l_flds[l_cnt] THEN CONTINUE FOR END IF
#      END FOR
      #170622-00040#1 mark(e)
      LET l_n = l_n + 1
   END WHILE
   FOR l_n = 1 TO l_flds.getlength()
      IF cl_null(l_flds[l_n]) THEN CONTINUE FOR END IF
      LET l_fld = l_flds[l_n]
#      SELECT COUNT(*) INTO l_cnt FROM gzgg_t   #170525-00105#1 mark
      SELECT COUNT(1) INTO l_cnt FROM gzgg_t    #170525-00105#1 add
       WHERE gzgg000 = l_gzgf000
         AND gzgg001 = l_fld
         AND gzgg006 = 'Y'
      IF l_cnt > 0 THEN
         LET l_gzgf.gzgf014 = l_gzgf.gzgf014 CLIPPED,l_flds[l_n],","
      END IF
   END FOR
   IF NOT cl_null(l_gzgf.gzgf014) THEN
      LET l_gzgf.gzgf014 = l_gzgf.gzgf014[1,LENGTH(l_gzgf.gzgf014)-1]
   ELSE
      LET l_gzgf.gzgf014 = " "
   END IF

   #檢查GROUP 欄位合理性 --end
   #檢查ORDER 欄位合理性 --start
   #IF NOT cl_null(l_gzgf015) THEN LET g_xgrid.order_field = l_gzgf015 END IF  
   LET l_gzgf.gzgf015 = DOWNSHIFT(l_gzgf.gzgf015)
   #LET lst_token = base.StringTokenizer.create(g_xgrid.order_field, ",")
   LET lst_token = base.StringTokenizer.create(l_gzgf.gzgf015, ",")
   LET l_n = 1
   CALL l_flds.clear()
   WHILE lst_token.hasMoreTokens()
      LET l_flds[l_n] = lst_token.nextToken()
      #170622-00040#1 mark(s)
#      FOR l_cnt = 1 TO l_flds.getlength()
#         IF l_cnt = l_n THEN CONTINUE FOR END IF
#         IF l_flds[l_n]=l_flds[l_cnt] THEN CONTINUE FOR END IF
#      END FOR
      #170622-00040#1 mark(e)
      LET l_n = l_n + 1
   END WHILE
   FOR l_n = 1 TO l_flds.getlength()
      IF cl_null(l_flds[l_n]) THEN CONTINUE FOR END IF
      LET l_fld = l_flds[l_n]
      LET l_fldstr = l_fld 
      IF l_fldstr.getIndexOf(" desc",1) > 0 THEN
         LET l_fld = cl_replace_str(l_fld," desc","")
         LET l_fldstr = " DESC"
      ELSE
         LET l_fldstr = NULL
      END IF
#      SELECT COUNT(*) INTO l_cnt FROM gzgg_t   #170525-00105#1 mark
      SELECT COUNT(1) INTO l_cnt FROM gzgg_t    #170525-00105#1 add
       WHERE gzgg000 = l_gzgf000
         AND gzgg001 = l_fld
         AND gzgg006 = 'Y'
         AND gzgg019 = '2'
      IF l_cnt > 0 THEN
         LET l_gzgf.gzgf015 = l_gzgf.gzgf015 CLIPPED,l_fld, l_fldstr CLIPPED,","
      END IF
   END FOR
   IF NOT cl_null(l_gzgf.gzgf015) THEN
      LET l_gzgf.gzgf015 = l_gzgf.gzgf015[1,LENGTH(l_gzgf.gzgf015)-1]
   ELSE
      LET l_gzgf.gzgf015 = " "
   END IF
   #檢查ORDER 欄位合理性 --end
    
   SELECT gzza003 INTO g_gzza003 FROM gzza_t
    WHERE gzza001 = g_prog
      AND gzzastus = "Y"   #141016 add    
   LET lc_sys_s = g_gzza003   
   LET lc_sys_s = UPSHIFT(lc_sys_s) CLIPPED
   IF lc_sys_s.subString(1,1) = 'C' THEN
      LET lc_sys_s = 'A',lc_sys_s.subString(2,lc_sys_s.getLength())
   END IF
   LET l_gzgh.gzgh016 = ""   #隱藏欄位
   #SELECT * INTO l_gzgf.* FROM gzgf_t WHERE gzgf000 = l_gzgf000
   #LET l_gzgh.gzgh001 = g_prog,YEAR(g_today) USING '<<<<',MONTH(g_today) USING '<<',DAY(g_today) USING '<<',
   #                  l_time[1,2],l_time[4,5],l_time[7,8]
   #LET l_gzgh.gzgh002 = l_gzgf000
  
   IF cl_null(g_xgrid.title1) THEN
      SELECT ooefl004 INTO l_gzgh.gzgh004 FROM ooefl_t WHERE ooeflent = g_enterprise AND ooefl001 = g_site AND ooefl002 = g_dlang
   ELSE
      LET l_gzgh.gzgh004 = g_xgrid.title1
   END IF
     
   IF cl_null(g_xgrid.title2) THEN
      #160921-00007#1 mark -(s)
#      SELECT gzdel003 INTO l_gzdel003 FROM gzdel_t 
#       WHERE gzdel001 = g_rep_code AND gzdel002 = g_dlang       
#   
#      IF NOT cl_null(l_gzdel003) THEN  
#         LET l_gzgh.gzgh005 = l_gzdel003 
#      #ELSE                             
#      #   LET l_gzgh.gzgh005 = l_gaz06  
#      END IF	
      #160921-00007#1 mark -(e)   
      #LET l_gzgh.gzgh005 = cl_get_progname(g_prog,g_dlang,1)   #160921-00007#1 add  #161227-00007#1 mark
      LET l_gzgh.gzgh005 = cl_gr_get_progname(g_rep_code,g_dlang,1)   #161227-00007#1 add
   ELSE
      LET l_gzgh.gzgh005 = g_xgrid.title2
   END IF

   LET l_gzgh.gzgh006 = NULL
#   LET l_gzgh.gzgh006 = g_user,"|",cl_get_username(g_user)                          #160330-00019#1 mark
   LET l_gzgh.gzgh006 = g_user,"|",cl_get_username(g_user),"|",FGL_GETENV("DGENV")   #160330-00019#1 add
#   DISPLAY "l_gzgh.gzgh006:",l_gzgh.gzgh006
#   DISPLAY "g_account:",g_account

   LET l_gzgh.gzgh044 = NULL  
   LET l_gzgh.gzgh044 = g_enterprise 

   LET l_gzgh.gzgh045 = NULL
   LET l_gzgh.gzgh045 = g_dbs
#   LET l_gzgh.gzgh049 = FGL_GETENV("PATH4RP") CLIPPED,"tmp"   #150916-00002#1 mark
#   LET l_gzgh.gzgh049 = cl_rpt_get_env_global("PATH4RP") CLIPPED,"tmp"   #150916-00002#1 add   #190219-00005#1 mark
   LET l_gzgh.gzgh049 = os.Path.join(cl_rpt_get_env_global("PATH4RP") CLIPPED,"tmp")   #190219-00005#1 add

   #允許列印方式權限控管
#   LET l_gzgh.gzgh050 = cl_gr_get_gzgo004("","","")   #150803-00008#2 add #150901-00021#1 mark      
   LET l_gzgh.gzgh050 = cl_rpt_get_gzgo004("","","")   #150901-00021#1 add    

   LET l_gzgf.gzgf007 = g_today
   #LET l_gzgf.gzgf016 = g_xgrid.skippage_field
   
   #取得營運中心圖檔
   LET l_logourl = os.Path.join(os.Path.join(FGL_GETENV("FGLASIP"),"pic"),"reptpic")
   LET l_logourl = os.Path.join(l_logourl,"pdf_logo_ds10.jpg")

    IF l_logourl.getIndexOf("http:",1) <> 1 THEN
       LET l_logourl = "http://",l_logourl
    END IF  

   #170301-00053#3 mark(s)
#   LET l_logopath = os.Path.join(os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("TOP"),"www"),"pic"),"reptpic"),"pdf_logo_ds10.jpg")
#   
#   IF NOT os.Path.exists(l_logopath) THEN
#      INITIALIZE g_errparam TO NULL
#      LET g_errparam.code = 'azz-00146'
#      LET g_errparam.extend = ''
#      LET g_errparam.popup = TRUE
#      CALL cl_err()
#
#      CALL cl_xg_view_addlog(cl_getmsg('azz-00146',g_rlang))
#      LET l_logopath = os.Path.join(os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("TOP"),"www"),"pic"),"reptpic"),"pdf_logo_ds10.jpg")
#      LET l_logourl = os.Path.join(os.Path.join(FGL_GETENV("FGLASIP"),"pic"),"reptpic")
#      LET l_logourl = os.Path.join(l_logourl,"pdf_logo_ds10.jpg")    
#   END IF
   #170301-00053#3 mark(e)

   LOCATE l_ooek003 IN MEMORY
   #logo對齊位置取azzi330對內表頭設定(YY)
#   DISPLAY "表頭代號:",l_gzgm003
   SELECT gzgj009 INTO l_gzgj009 FROM gzgj_t WHERE gzgj001 = l_gzgm003 AND gzgj002 = g_dlang AND gzgjstus = "Y"
      AND gzgj010 = '1'   #150611-00032#4 add
   IF cl_null(l_gzgj009) THEN LET l_gzgj009 = "1" END IF   #預設給1:靠左    
   
   #150904-00005#1 add(s)
   #列印條件
   SELECT gzgj009 INTO l_pri_con FROM gzgj_t 
    WHERE gzgj001 = l_gzgm003 AND gzgj002 = g_dlang AND gzgjstus = "Y" AND gzgj010 = '2'
   
   IF l_pri_con = "Y" THEN
      IF NOT cl_null(g_rep_wcchp) THEN LET l_gzgh.gzgh008 = g_rep_wcchp END IF
      
      IF NOT cl_null(g_xgrid.condition) THEN
         LET l_gzgh.gzgh008 = g_xgrid.condition
      END IF
   ELSE
      LET l_pri_con = "N"
   END IF   
   LET l_gzgh.gzgh008 = l_pri_con CLIPPED,"|",cl_getmsg("lib-00391",g_lang) CLIPPED,":",l_gzgh.gzgh008   #150904-00005#4 add      
   #150904-00005#1 add(e)   
   
   SELECT ooek003 INTO l_ooek003 FROM ooek_t
    WHERE ooekstus = "Y" AND ooekent = g_enterprise AND ooek001 = g_site AND ooek002 = g_dlang
   
   IF l_ooek003 IS NULL THEN
#      LET l_logourl = os.Path.join(os.Path.join(FGL_GETENV("FGLASIP"),"pic"),"reptpic")
#      LET l_logourl = os.Path.join(l_logourl,"no_logo.png")
#      LET l_gzgh.gzgh009 = l_gzgj009,";",l_logourl
#      IF g_user = '04682' THEN
         LET l_gzgh.gzgh009 = l_gzgj009,";","no_logo.png"
#      END IF   
   ELSE
#      LET l_gzgh.gzgh009 = l_gzgj009,";","SELECT ooek003 FROM ",g_dbs CLIPPED,".ooek_t WHERE ooekstus = 'Y' ",
#                           " AND ooekent =", g_enterprise CLIPPED," AND ooek001 = '",g_site CLIPPED,"' AND ooek002='",g_dlang CLIPPED,"'"
#      IF g_user = '04682' THEN
         LET l_logopath = g_enterprise
         LET l_logopath = g_dbs CLIPPED,"_",l_logopath CLIPPED,"_",g_site CLIPPED,"_",g_dlang CLIPPED,".jpg"
         LET l_gzgh.gzgh009 = l_gzgj009,";",l_logopath
#      END IF   
   END IF  
#   DISPLAY "l_gzgh.gzgh009:",l_gzgh.gzgh009   
   
   IF cl_null(l_gzgh.gzgh012) THEN LET l_gzgh.gzgh012 = 'A' END IF
   LET l_gzgh.gzgh013 = l_gzgh013
   LET l_gzgh.gzgh001 = l_certid
#   LET l_gzgh.gzgh007 = DATE(CURRENT)   #161117-00037#1 mark
   LET l_gzgh.gzgh007 = g_ldate          #161117-00037#1 add
   LET l_gzgh.gzgh015 = g_roles   
   LET l_gzgh.gzgh017 = g_xgrid.headerinfo1
   LET l_gzgh.gzgh018 = g_xgrid.headerinfo2
   LET l_gzgh.gzgh019 = g_xgrid.headerinfo3
   LET l_gzgh.gzgh020 = g_xgrid.headerinfo4
   LET l_gzgh.gzgh021 = g_xgrid.headerinfo5
   CALL cl_xg_fix_info(g_xgrid.footerinfo1,g_xgrid.footerinfo2,g_xgrid.footerinfo3,g_xgrid.footerinfo4,g_xgrid.footerinfo5)
      RETURNING g_xgrid.footerinfo1,g_xgrid.footerinfo2,g_xgrid.footerinfo3,g_xgrid.footerinfo4,g_xgrid.footerinfo5
   LET l_gzgh.gzgh022 = g_xgrid.footerinfo1
   LET l_gzgh.gzgh023 = g_xgrid.footerinfo2   
   LET l_gzgh.gzgh024 = g_xgrid.footerinfo3   #補表尾資訊3 #180919-00051#1 add   
   LET l_gzgh.gzgh025 = g_xgrid.footerinfo4
   LET l_gzgh.gzgh026 = g_xgrid.footerinfo5

   CALL cl_xg_apr_str() RETURNING l_oojj005_desc
   LET l_gzgh.gzgh029 = l_oojj005_desc   #簽核字串

#   CALL cl_gr_get_numfmt_str() RETURNING l_str #數值欄位格式   #150901-00021#1 mark
   CALL cl_rpt_get_numfmt_str() RETURNING l_str #數值欄位格式   #150901-00021#1 add
   LET l_gzgh.gzgh043 = l_str                  #數值屬性的小數位數
   
   LET l_gzgh.gzgh042 = g_rlang 

   LET l_gzgh.gzgh037 = 'N'

#   CALL cl_xg_view_updatenull(l_gzgh.gzgh003)    #160921-00007#1 mark
    
   LET l_gzgh.gzgh027 = g_xgrid.dynamic_title #動態標題
   
   CALL cl_xg_maillist(l_gzgh.*) RETURNING l_gzgh.*
#   IF g_bgjob = "Y" THEN   #150513 mark
#150803-00008#2 mark(s)
#      LET l_gzgh.gzgh030 = NULL
#      LET l_gzgh.gzgh031 = NULL
#      LET l_gzgh.gzgh032 = NULL
#      
#      IF g_schedule_d.getLength() > 0 THEN
#         FOR li_idx = 1 TO g_schedule_d.getLength()
#            LET g_schedule_d[li_idx].gzpe009 = s_aooi350_get_user_contact(g_schedule_d[li_idx].gzpe004,'4')   #150611-00032#4 add 
#            CASE g_schedule_d[li_idx].gzpe007
#               WHEN "1"
#                  LET l_gzgh.gzgh030 = l_gzgh.gzgh030,";",g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010
#               WHEN "2"
#                  LET l_gzgh.gzgh031 = l_gzgh.gzgh031,";",g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010
#               WHEN "3"
#                  LET l_gzgh.gzgh032 = l_gzgh.gzgh032,";",g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010               
#            END CASE
#         END FOR
#         LET l_str = l_gzgh.gzgh030 CLIPPED
#         LET l_str = l_str.trim()
#         LET l_gzgh.gzgh030 = l_str.substring(2,LENGTH(l_str))
#         LET l_str = l_gzgh.gzgh031 CLIPPED
#         LET l_str = l_str.trim()        
#         LET l_gzgh.gzgh031 = l_str.substring(2,LENGTH(l_str)) 
#         LET l_str = l_gzgh.gzgh032 CLIPPED
#         LET l_str = l_str.trim()                  
#         LET l_gzgh.gzgh032 = l_str.substring(2,LENGTH(l_str))                  
#      END IF
#150803-00008#2 mark(e)      
#      DISPLAY "l_gzgh.gzgh030: ",l_gzgh.gzgh030 
#      DISPLAY "l_gzgh.gzgh031: ",l_gzgh.gzgh031
#      DISPLAY "l_gzgh.gzgh032: ",l_gzgh.gzgh032       
#   END IF   #150513 mark
   
   INSERT INTO gzgh_t(gzgh001,gzgh002,gzgh003,gzgh004,gzgh005,gzgh006,gzgh007,gzgh008,gzgh009,gzgh010,
                      gzgh011,gzgh012,gzgh013,gzgh014,gzgh015,gzgh016,gzgh017,gzgh018,gzgh019,gzgh020,
                      gzgh021,gzgh022,gzgh023,gzgh024,gzgh025,gzgh026,gzgh027,gzgh028,gzgh029,gzgh030,
                      gzgh031,gzgh032,gzgh033,gzgh034,gzgh035,gzgh036,gzgh037,gzgh038,gzgh039,gzgh040,
#                      gzgh041,gzgh042,gzgh043,gzgh044,gzgh045,gzgh046,gzgh047,gzgh048,gzgh049,gzgh050) #150803-00008#2 add gzgh050   #150803-00008#2 mark
#                      gzgh041,gzgh042,gzgh043,gzgh044,gzgh045,gzgh046,gzgh047,gzgh048,gzgh049,gzgh050,gzgh051) #150803-00008#2 add   #190107-00075#8 mark
                      gzgh041,gzgh042,gzgh043,gzgh044,gzgh045,gzgh046,gzgh047,gzgh048,gzgh049,gzgh050,gzgh051,gzgh052)   #190107-00075#8 add gzgh052
               VALUES(l_gzgh.gzgh001,l_gzgh.gzgh002,l_gzgh.gzgh003,l_gzgh.gzgh004,l_gzgh.gzgh005,
                      l_gzgh.gzgh006,l_gzgh.gzgh007,l_gzgh.gzgh008,l_gzgh.gzgh009,l_gzgh.gzgh010,
                      l_gzgh.gzgh011,l_gzgh.gzgh012,l_gzgh.gzgh013,l_gzgh.gzgh014,l_gzgh.gzgh015,
                      l_gzgh.gzgh016,l_gzgh.gzgh017,l_gzgh.gzgh018,l_gzgh.gzgh019,l_gzgh.gzgh020,
                      l_gzgh.gzgh021,l_gzgh.gzgh022,l_gzgh.gzgh023,l_gzgh.gzgh024,l_gzgh.gzgh025,
                      l_gzgh.gzgh026,l_gzgh.gzgh027,l_gzgh.gzgh028,l_gzgh.gzgh029,l_gzgh.gzgh030,
                      l_gzgh.gzgh031,l_gzgh.gzgh032,l_gzgh.gzgh033,l_gzgh.gzgh034,l_gzgh.gzgh035,
                      l_gzgh.gzgh036,l_gzgh.gzgh037,l_gzgh.gzgh038,l_gzgh.gzgh039,l_gzgh.gzgh040,
                      l_gzgh.gzgh041,l_gzgh.gzgh042,l_gzgh.gzgh043,l_gzgh.gzgh044,l_gzgh.gzgh045,
#                      l_gzgh.gzgh046,l_gzgh.gzgh047,l_gzgh.gzgh048,l_gzgh.gzgh049,l_gzgh.gzgh050)   #150803-00008#2 add gzgh050   #150803-00008#2 mark           
#                      l_gzgh.gzgh046,l_gzgh.gzgh047,l_gzgh.gzgh048,l_gzgh.gzgh049,l_gzgh.gzgh050,l_gzgh.gzgh051)   #150803-00008#2 add #190107-00075#8 mark
                      l_gzgh.gzgh046,l_gzgh.gzgh047,l_gzgh.gzgh048,l_gzgh.gzgh049,l_gzgh.gzgh050,l_gzgh.gzgh051,l_gzgh.gzgh052)   #190107-00075#8 add gzgh052
   
   IF STATUS THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = STATUS
      LET g_errparam.extend = 'ins gdq:'
      LET g_errparam.popup = TRUE
      CALL cl_err()

      CALL cl_xg_view_addlog("INSERT INTO gzgh_t ERROR:"||STATUS)
      RETURN
   END IF

   LET l_gzgh.gzgh041 = l_gzgh041
   UPDATE gzgh_t SET gzgh041 = l_gzgh.gzgh041 WHERE gzgh001 = l_gzgh.gzgh001 #150402 add 
    
   FREE l_gzgh.gzgh041 

   #150709-00014#1 add(s)
   LET g_msgparam.rpt_code = l_certid
   LET g_msgparam.rpt_url = l_url
#   LET l_file_name = FGL_GETENV("XTRAGRIDIP") CLIPPED,"temp/"   #150916-00002#1 mark 
   LET l_file_name = cl_rpt_get_env_global("XTRAGRIDIP") CLIPPED,"temp/"   #150916-00002#1 add   #170209-00022#1 mark    #170309-00023#1 remark
#   LET l_file_name = cl_rpt_get_env_global(g_ip_type) CLIPPED,"temp/"                             #170209-00022#1 add   #170309-00023#1 mark
   LET l_file_name = l_file_name CLIPPED ,g_site CLIPPED,"_",g_prog CLIPPED,"_",g_user CLIPPED,".pdf"
   LET g_msgparam.rpt_file = l_file_name   
   #150709-00014#1 add(e)
   
   LET l_ostype = FGL_GETENV("ostype") #151216-00001#1 add
   DISPLAY "l_ostype:",l_ostype
   
#   IF g_bgjob != 'Y' THEN   #151216-00001#1 add
   IF cl_null(g_schedule_id) OR g_schedule_id = 0 THEN                      #151216-00001#1 add
      CALL ui.Interface.frontCall("standard","feinfo", "ostype", l_ostype)  #150508-00001#1 for MAC add
   END IF                                                                   #151216-00001#1 add

   #150803-00008#4 add
   #151216-00001#1 add(s)
   IF NOT cl_null(g_schedule_id) THEN
      CALL cl_xg_update_gzpe009()
   END IF
   #151216-00001#1 add(e)
   
   #150709-00014#1 add(s)
   #信件主旨/信件內文若為'(系統預設)'或NULL,要置換成'default'
   IF g_schedule_d.getLength() > 0 THEN
      FOR l_i = 1 TO g_schedule_d.getLength()
         #信件主旨
         IF cl_null(g_schedule_d[l_i].gzpe005) THEN
            LET g_schedule_d[l_i].gzpe005 = 'default'
         ELSE
            LET l_gzze003 = g_schedule_d[l_i].gzpe005 CLIPPED
            SELECT COUNT(gzze001) INTO l_cnt FROM gzze_t 
             WHERE gzze001 = 'lib-00236' AND gzze003 = l_gzze003
             IF l_cnt > 0 THEN LET g_schedule_d[l_i].gzpe005 = 'default' END IF
         END IF
         
         #信件內文
         IF cl_null(g_schedule_d[l_i].gzpe006) THEN
            LET g_schedule_d[l_i].gzpe006 = 'default'
         ELSE
            LET l_gzze003 = g_schedule_d[l_i].gzpe006 CLIPPED
            SELECT COUNT(gzze001) INTO l_cnt FROM gzze_t 
             WHERE gzze001 = 'lib-00236' AND gzze003 = l_gzze003
             IF l_cnt > 0 THEN LET g_schedule_d[l_i].gzpe006 = 'default' END IF
         END IF         
      END FOR
   END IF
   DISPLAY "g_url: ", g_url       
   IF cl_msgcentre_report() THEN      
      DISPLAY "訊息中心"
#      UPDATE gzgh_t SET gzgh048 = "axmr001-00544-B4NB|1"
#       WHERE gzgh001 = l_gzgh.gzgh001
#      LET g_bgjob = 'Y'
#      LET g_schedule_id = "axmr001-00544-B4NB" 
#      LET g_schedule_seq = "1"
      LET ls_run_cmd = "r.r awsp300 ",l_gzgh.gzgh001," ", g_schedule_id," ",g_schedule_seq
      CALL cl_cmdrun_openpipe("r.r awsp300", ls_run_cmd, FALSE) RETURNING lb_run_result,ls_err_msg
#      IF (FGL_GETENV("GUIMODE") != "GWC" OR FGL_GETENV("GUIMODE") IS NULL) AND l_ostype.toUpperCase() = "WINDOWS" THEN   #190219-00005#1 mark
      IF (FGL_GETENV("GUIMODE") != "GWC" OR FGL_GETENV("GUIMODE") IS NULL) AND l_ostype.toUpperCase() = "WINDOWS" #190219-00005#1 add
          AND g_defBrowser <> "Y" THEN   #190219-00005#1 add
#         LET l_url = 'iexplore -noframemerging '," \"",l_url,"\""         #180919-00051#1 mark
#         CALL ui.Interface.frontCall("standard","shellexec",l_url ,res)   #180919-00051#1 mark         
         #XP且登入簡中時，使用Firefox，XP其他語系使用Chrome，非XP使用IE
         IF l_version = "Windows XP" THEN
            IF g_lang = "zh_CN" THEN
               LET l_url = "cmd.exe /C start firefox -new-window ",l_url
               CALL ui.Interface.frontCall("standard","execute", [l_url,1] ,[res])
               IF res = 0 THEN
                  INITIALIZE g_errparam TO NULL
                  LET g_errparam.code = "lib-00449"   #請確認是否正確安裝Firefox瀏覽器！
                  LET g_errparam.extend = ''
                  LET g_errparam.replace[1] = "Firefox"
                  LET g_errparam.popup = TRUE
                  CALL cl_err()
               END IF
            ELSE
               LET l_url = "cmd.exe /C start chrome /new-window ",l_url
               CALL ui.Interface.frontCall("standard","execute", [l_url,1] ,[res])
               IF res = 0 THEN
                  INITIALIZE g_errparam TO NULL
                  LET g_errparam.code = "lib-00449"   #請確認是否正確安裝Chrome瀏覽器！
                  LET g_errparam.extend = ''
                  LET g_errparam.replace[1] = "Chrome"
                  LET g_errparam.popup = TRUE
                  CALL cl_err()
               END IF
            END IF
         ELSE
            LET l_url = 'iexplore -noframemerging '," \"",l_url,"\""
            CALL ui.Interface.frontCall("standard","shellexec",l_url ,res)
         END IF
         #180919-00051#1 add(e)
      ELSE      
         #260327-00001#1 add(s)
         IF FGL_GETENV("T100NANAWEB") = 'Y' THEN
            DISPLAY "T100NaNaWeb Run Xg Report"
            #回傳網址處理
            LET la_param.prog = "adzp902"
            LET la_param.param[1] = "open_t100_report_xg"
            LET la_param.param[2] = l_url
            LET lc_str = util.JSON.stringify( la_param )   
            CALL cl_cmdrun(lc_str)
         ELSE  
            #直接開啟
            CALL ui.Interface.frontCall("standard","launchurl",l_url ,res)
         END IF
         #260327-00001#1 add(e)
         #CALL ui.Interface.frontCall("standard","launchurl",l_url ,res) #260327-00001#1 mark
      END IF
   ELSE   
   #150709-00014#1 add(e)
      IF g_bgjob = 'Y' THEN
#         CALL cl_xg_view_bgjob(l_url) #背景作業
         LET ls_run_cmd = "r.r awsp300 ",l_gzgh.gzgh001," ", g_schedule_id," ",g_schedule_seq
         CALL cl_cmdrun_openpipe("r.r awsp300", ls_run_cmd, FALSE) RETURNING lb_run_result,ls_err_msg
      
         LET g_run_cmd = ls_run_cmd
         LET g_run_result = lb_run_result
         LET g_err_msg = ls_err_msg
         CALL cl_xg_view_addlog(g_err_msg)   #190219-00005#1 add
      ELSE
         #150508-00001#1 for MAC add(s)      
#         IF FGL_GETENV("GUIMODE") != "GWC" AND l_ostype.toUpperCase() = "WINDOWS" THEN  #150508-00001#3 mark
#         IF (FGL_GETENV("GUIMODE") != "GWC" OR FGL_GETENV("GUIMODE") IS NULL) AND l_ostype.toUpperCase() = "WINDOWS" THEN #150508-00001#3 add   #190219-00005#1 mark
         IF (FGL_GETENV("GUIMODE") != "GWC" OR FGL_GETENV("GUIMODE") IS NULL) AND l_ostype.toUpperCase() = "WINDOWS"       #190219-00005#1 add
            AND g_defBrowser <> "Y" THEN   #190219-00005#1 add
#            LET l_url = 'iexplore -noframemerging '," \"",l_url,"\""         #180919-00051#1 mark
#            CALL ui.Interface.frontCall("standard","shellexec",l_url ,res)   #180919-00051#1 mark
            #180919-00051#1 add(s)
            #XP且登入簡中時，使用Firefox，XP其他語系使用Chrome，非XP使用IE
            IF l_version = "Windows XP" THEN
               IF g_lang = "zh_CN" THEN
                  LET l_url = "cmd.exe /C start firefox -new-window ",l_url
                  CALL ui.Interface.frontCall("standard","execute", [l_url,1] ,[res])
                  IF res = 0 THEN
                     INITIALIZE g_errparam TO NULL
                     LET g_errparam.code = "lib-00449"   #請確認是否正確安裝Firefox瀏覽器！
                     LET g_errparam.extend = ''
                     LET g_errparam.replace[1] = "Firefox"
                     LET g_errparam.popup = TRUE
                     CALL cl_err()
                  END IF
               ELSE
                  LET l_url = "cmd.exe /C start chrome /new-window ",l_url
                  CALL ui.Interface.frontCall("standard","execute", [l_url,1] ,[res])
                  IF res = 0 THEN
                     INITIALIZE g_errparam TO NULL
                     LET g_errparam.code = "lib-00449"   #請確認是否正確安裝Chrome瀏覽器！
                     LET g_errparam.extend = ''
                     LET g_errparam.replace[1] = "Chrome"
                     LET g_errparam.popup = TRUE
                     CALL cl_err()
                  END IF
               END IF
            ELSE
               LET l_url = 'iexplore -noframemerging '," \"",l_url,"\""
               CALL ui.Interface.frontCall("standard","shellexec",l_url ,res)
            END IF
            #180919-00051#1 add(e)
         ELSE   
            #260327-00001#1 add(s)
            IF FGL_GETENV("T100NANAWEB") = 'Y' THEN
               DISPLAY "T100NaNaWeb Run Xg Report"
               #回傳網址處理
               LET la_param.prog = "adzp902"
               LET la_param.param[1] = "open_t100_report_xg"
               LET la_param.param[2] = l_url
               LET lc_str = util.JSON.stringify( la_param )   
               CALL cl_cmdrun(lc_str)
            ELSE  
               #直接開啟
               CALL ui.Interface.frontCall("standard","launchurl",l_url ,res)
            END IF
            #260327-00001#1 add(e)         
            #CALL ui.Interface.frontCall("standard","launchurl",l_url ,res) #260327-00001#1 mark
         END IF
         #150508-00001#1 for MAC add(e) 
      
         #150508-00001#1 for MAC mark(s) 
#         IF FGL_GETENV("GUIMODE") = "GWC" THEN
#            CALL ui.Interface.frontCall("standard","launchurl",l_url ,res)
#         ELSE      
#            LET l_url = 'iexplore -noframemerging '," \"",l_url,"\""
#            CALL ui.Interface.frontCall("standard","shellexec",l_url ,res)
#         END IF         
         #150508-00001#1 for MAC mark(e) 
      END IF
   END IF #150709-00014#1 add
   
   #190219-00005#1 add(s)
   LET g_err_msg = g_err_msg,"\nTemplate gzgf000:",l_tmps[1].gzgf000,",gzgf002:",l_tmps[1].gzgf002,",gzgf004:",l_tmps[1].gzgf004,
                   ",gzgf005:",l_tmps[1].gzgf005
   DISPLAY g_err_msg
   #190219-00005#1 add(e)
   CALL cl_xg_view_logfile()
   INITIALIZE g_xgrid.* TO NULL

END FUNCTION

################################################################################
# Descriptions...: 設定XtraGrid的URL字串
# Memo...........:
# Usage..........: cl_xg_set_url()
#                  RETURNING 回传参数
# Return code....: l_certid
#                : l_url
#                : l_url_param
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_set_url()
   DEFINE  l_rep_db,l_instance,l_server    STRING
   DEFINE  l_rep_db_pw,l_str               STRING
   DEFINE  l_rep_db_user,l_sqlserver_ip    STRING
   DEFINE  l_certid,l_url,l_db_type        STRING
   DEFINE  lc_azz06                        LIKE type_t.chr1 
   DEFINE  lc_zz06                         LIKE type_t.chr1 
   DEFINE  li_rptlimit                     LIKE type_t.num10
   DEFINE  crip,ms_codeset                 STRING      ## Report Server IP ##
   DEFINE  certid                          LIKE type_t.num5
   #DEFINE  l_plant                         STRING   #141016 mark
   DEFINE  lc_aza49                        LIKE type_t.chr1
   DEFINE  l_azp20                         LIKE type_t.chr20
   DEFINE  l_url_param                     STRING        #存在zax42的網址參數
   DEFINE  l_time                          STRING
   DEFINE  l_sb                            base.StringBuffer

   LET l_rep_db = g_dbs CLIPPED
   LET l_rep_db_pw = ""
   LET l_instance = ""
   LET l_db_type=cl_db_get_database_type()

   LET lc_azz06 ='Y'                          #暫用  
   CASE
      WHEN l_db_type="ORA"
         LET l_str = "dbi.database.", g_report CLIPPED, ".password"  #140910 mark
         #LET l_str = "dbi.database.", g_rep_db CLIPPED, ".password" #140910 add
         LET l_rep_db_pw = fgl_getresource(l_str) CLIPPED
         LET l_instance = fgl_getenv('ORACLE_SID') CLIPPED

         CALL cl_xg_trans_ascii(l_rep_db_user.trim()) RETURNING l_str  ## DB ##
         LET l_rep_db_user = l_str.trim()
         LET l_str = ""
         CALL cl_xg_trans_ascii(l_rep_db_pw.trim()) RETURNING l_str ## DB_PW ##
         LET l_rep_db_pw = l_str.trim()
         LET l_str = ""
         CALL cl_xg_trans_ascii(l_instance.trim()) RETURNING l_str  ## INSTANCE ##
         LET l_instance = l_str.trim()

         #LET g_asp_param.t = l_rep_db
         #LET g_asp_param.t1 = l_rep_db_pw
#         IF g_ip_type = 'XTRAGRIDIP' THEN   #170209-00022#1 add   #170309-00023#1 mark
            LET l_sqlserver_ip = cl_xg_get_serverip("FGLASIP")            
         #170309-00023#1 mark(s)            
#         #170209-00022#1 add(s)         
#         ELSE
#            LET l_sqlserver_ip = os.Path.join(cl_rpt_get_env_global("XTRAGRIDIP"),FGL_GETENV("ZONE"))
#         END IF
#         #170209-00022#1 add(e)
         #170309-00023#1 mark(e)        
         CALL cl_xg_trans_ascii(l_sqlserver_ip.trim()) RETURNING l_str
         LET l_sqlserver_ip = l_str.trim()
         LET l_url_param = "user=",l_rep_db_user,"&pswd=",l_rep_db_pw,"&sid=",l_instance,"&ip=",l_sqlserver_ip
   END CASE
   
   #CALL cl_get_plant() RETURNING l_plant   #141016 mark

   LET l_time = CURRENT YEAR TO FRACTION(3)  #時間截記
   LET l_sb = base.StringBuffer.create()
   CALL l_sb.append(l_time)
   CALL l_sb.replace(":","",0)
   CALL l_sb.replace(" ","",0)
   CALL l_sb.replace("-","",0)
   CALL l_sb.replace(".","",0)
   LET l_time = l_sb.toString()  
#   LET crip = FGL_GETENV("XTRAGRIDIP")   #Report Server IP   #150916-00002#1 mark 
  LET crip = cl_rpt_get_env_global("XTRAGRIDIP")  #Report Server IP   #150916-00002#1 add   #170209-00022#1 mark     #170309-00023#1 remark
#   LET crip = cl_rpt_get_env_global(g_ip_type)  #Report Server IP                             #170209-00022#1 add   #170309-00023#1 mark
   LET certid = FGL_GETPID()
   LET l_certid = certid CLIPPED,l_time
   #DISPLAY "certid=",l_certid
   #l_certid由PID+年月日+時間(時分秒毫秒)，例如:13745_20160415_060851749
   LET l_certid = l_certid.trim() 
   #LET l_url = crip CLIPPED,"Default.aspx"                       ##150916-00002#1 mark
   LET l_url = os.Path.join(crip,"Default.aspx")                  ##150916-00002#1 mark

   LET l_url = l_url CLIPPED,"?report_sn=",l_certid CLIPPED
   
   LET l_url = l_url CLIPPED,"&defBrowser=",g_defBrowser CLIPPED  #190219-00005#1 add #預設瀏覽器參數

   RETURN l_certid,l_url,l_url_param
END FUNCTION

################################################################################
# Descriptions...: 取得ERP主機的IP
# Memo...........:
# Usage..........: cl_xg_get_serverip()
#                  RETURNING 回传参数
# Return code....: l_ip   
# Date & Author..: 140703 By 02716
# Modify.........: 2014/12/18 By 04010
################################################################################
PUBLIC FUNCTION cl_xg_get_serverip(p_var)
   DEFINE p_var STRING
   DEFINE l_ip  LIKE type_t.chr100
   DEFINE l_n1,l_n2  LIKE type_t.num5
   DEFINE l_str  STRING

#   LET l_str = FGL_GETENV("FGLASIP")
   IF p_var = "XTRAGRIDIP" THEN
      LET l_str = cl_rpt_get_env_global("XTRAGRIDIP") CLIPPED
   ELSE
      LET l_str = FGL_GETENV(p_var)   
   END IF
   LET l_n1 = l_str.getIndexOf("http://",1)
   LET l_n2 = l_str.getIndexOf("/",l_n2+8)
   LET l_ip = l_str
   LET l_ip = l_ip[l_n1+7,l_n2-1]
   RETURN l_ip
END FUNCTION

################################################################################
# Descriptions...: #報表尾資訊整理與對齊
# Memo...........:
# Usage..........: cl_xg_fix_info(p_info1,p_info2,p_info3,p_info4,p_info5)
#                  RETURNING 回传参数
# Input parameter: p_info1
#                : p_info2
#                : p_info3
#                : p_info4
#                : p_info5
# Return code....: p_info1
#                : p_info2
#                : p_info3
#                : p_info4
#                : p_info5
# Date & Author..: 141017 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_fix_info(p_info1,p_info2,p_info3,p_info4,p_info5)
   DEFINE p_info1    LIKE type_t.chr1000
   DEFINE p_info2    LIKE type_t.chr1000
   #DEFINE p_info3    LIKE type_t.chr1000   #141016 mark
   DEFINE p_info3    LIKE type_t.chr500   #141016 add
   DEFINE p_info4    LIKE type_t.chr1000
   DEFINE p_info5    LIKE type_t.chr1000
   DEFINE l_str      STRING
   DEFINE l_i,l_n,i  LIKE type_t.num5
   DEFINE l_fld      LIKE type_t.chr1000
   DEFINE l_len      LIKE type_t.num5
   DEFINE l_len1     LIKE type_t.num5
   DEFINE l_len2     LIKE type_t.num5
   DEFINE l_len3     LIKE type_t.num5
   DEFINE l_len4     LIKE type_t.num5
   DEFINE l_len5     LIKE type_t.num5
   DEFINE lst_token  base.StringTokenizer
   DEFINE ls_token   STRING
   DEFINE l_flag1    LIKE type_t.chr1
   DEFINE l_flag2    LIKE type_t.chr1
   DEFINE l_flag3    LIKE type_t.chr1
   DEFINE l_flag4    LIKE type_t.chr1
   DEFINE l_flag5    LIKE type_t.chr1
   DEFINE l_flds     DYNAMIC ARRAY OF RECORD
                        section1 LIKE type_t.chr1000,
                        section2 LIKE type_t.chr1000,
                        section3 LIKE type_t.chr1000,
                        section4 LIKE type_t.chr1000,
                        section5 LIKE type_t.chr1000
                     END RECORD
   DEFINE l_oojj005  LIKE oojj_t.oojj005 #簽核代號
   DEFINE l_sql      STRING   
   
   LET l_flag1 = 'N'
   LET l_flag2 = 'N'
   LET l_flag3 = 'N'
   LET l_flag4 = 'N'
   LET l_flag5 = 'N'
   #拆解 footerinfo1 
   IF NOT cl_null(p_info1) THEN
      LET lst_token = base.StringTokenizer.create(p_info1, "|")
      LET l_n = 1
      WHILE lst_token.hasMoreTokens()
         LET l_fld = lst_token.nextToken()
         CASE l_n
            WHEN 1 LET l_flds[1].section1 = l_fld
            WHEN 2 LET l_flds[1].section2 = l_fld
            WHEN 3 LET l_flds[1].section3 = l_fld
            WHEN 4 LET l_flds[1].section4 = l_fld
            WHEN 5 LET l_flds[1].section5 = l_fld
         END CASE
         LET l_flag1 = 'Y'
         LET l_n = l_n + 1
         IF l_n > 5 THEN EXIT WHILE END IF
      END WHILE
   END IF
   
   #拆解 footerinfo2 
   IF NOT cl_null(p_info2) THEN
      LET lst_token = base.StringTokenizer.create(p_info2, "|")
      LET l_n = 1
      WHILE lst_token.hasMoreTokens()
         LET l_fld = lst_token.nextToken()
         CASE l_n
            WHEN 1 LET l_flds[2].section1 = l_fld
            WHEN 2 LET l_flds[2].section2 = l_fld
            WHEN 3 LET l_flds[2].section3 = l_fld
            WHEN 4 LET l_flds[2].section4 = l_fld
            WHEN 5 LET l_flds[2].section5 = l_fld
         END CASE
         LET l_n = l_n + 1
         LET l_flag2 = 'Y'
         IF l_n > 5 THEN EXIT WHILE END IF
      END WHILE
   END IF

   #拆解 footerinfo3
   IF NOT cl_null(p_info3) THEN
      LET lst_token = base.StringTokenizer.create(p_info3, "|")
      LET l_n = 1
      WHILE lst_token.hasMoreTokens()
         LET l_fld = lst_token.nextToken()
         CASE l_n
            WHEN 1 LET l_flds[3].section1 = l_fld
            WHEN 2 LET l_flds[3].section2 = l_fld
            WHEN 3 LET l_flds[3].section3 = l_fld
            WHEN 4 LET l_flds[3].section4 = l_fld
            WHEN 5 LET l_flds[3].section5 = l_fld
         END CASE
         LET l_n = l_n + 1
         LET l_flag3 = 'Y'
         IF l_n > 5 THEN EXIT WHILE END IF
      END WHILE
   END IF

   #拆解 footerinfo4
   IF NOT cl_null(p_info4) THEN
      LET lst_token = base.StringTokenizer.create(p_info4, "|")
      LET l_n = 1
      WHILE lst_token.hasMoreTokens()
         LET l_fld = lst_token.nextToken()
         CASE l_n
            WHEN 1 LET l_flds[4].section1 = l_fld
            WHEN 2 LET l_flds[4].section2 = l_fld
            WHEN 3 LET l_flds[4].section3 = l_fld
            WHEN 4 LET l_flds[4].section4 = l_fld
            WHEN 5 LET l_flds[4].section5 = l_fld
         END CASE
         LET l_n = l_n + 1
         LET l_flag4 = 'Y'
         IF l_n > 5 THEN EXIT WHILE END IF
      END WHILE
   END IF

   #拆解 footerinfo5
   IF NOT cl_null(p_info5) THEN
      LET lst_token = base.StringTokenizer.create(p_info5, "|")
      LET l_n = 1
      WHILE lst_token.hasMoreTokens()
         LET l_fld = lst_token.nextToken()
         CASE l_n
            WHEN 1 LET l_flds[5].section1 = l_fld
            WHEN 2 LET l_flds[5].section2 = l_fld
            WHEN 3 LET l_flds[5].section3 = l_fld
            WHEN 4 LET l_flds[5].section4 = l_fld
            WHEN 5 LET l_flds[5].section5 = l_fld
         END CASE
         LET l_n = l_n + 1
         LET l_flag5 = 'Y'
         IF l_n > 5 THEN EXIT WHILE END IF
      END WHILE
   END IF

   LET l_len1= 0 
   LET l_len2= 0 
   LET l_len3= 0 
   LET l_len4= 0 
   LET l_len5= 0 
   #找出每個Section最寬的
   FOR i = 1 TO 5
      LET l_len = FGL_WIDTH(l_flds[i].section1)
      IF (l_len1 = 0) OR (l_len > l_len1) THEN LET l_len1=l_len END IF
      LET l_len = FGL_WIDTH(l_flds[i].section2)
      IF (l_len2 = 0) OR (l_len > l_len2) THEN LET l_len2=l_len END IF
      LET l_len = FGL_WIDTH(l_flds[i].section3)
      IF (l_len3 = 0) OR (l_len > l_len3) THEN LET l_len3=l_len END IF
      LET l_len = FGL_WIDTH(l_flds[i].section4)
      IF (l_len4 = 0) OR (l_len > l_len4) THEN LET l_len4=l_len END IF
      LET l_len = FGL_WIDTH(l_flds[i].section5)
      IF (l_len5 = 0) OR (l_len > l_len5) THEN LET l_len5=l_len END IF
   END FOR
   #填空
   FOR i = 1 TO 5
      LET l_len = FGL_WIDTH(l_flds[i].section1)
      IF l_len1 > l_len AND l_len > 0 THEN
         FOR l_i = 1 TO l_len1-l_len
            LET l_flds[i].section1=l_flds[i].section1,' '
         END FOR
      END IF
      LET l_len = FGL_WIDTH(l_flds[i].section2)
      IF l_len2 > l_len AND l_len > 0 THEN
         FOR l_i = 1 TO l_len2-l_len
            LET l_flds[i].section2=l_flds[i].section2,' '
         END FOR
      END IF
      LET l_len = FGL_WIDTH(l_flds[i].section3)
      IF l_len3 > l_len AND l_len > 0 THEN
         FOR l_i = 1 TO l_len3-l_len
            LET l_flds[i].section3=l_flds[i].section3,' '
         END FOR
      END IF
      LET l_len = FGL_WIDTH(l_flds[i].section4)
      IF l_len4 > l_len AND l_len > 0 THEN
         FOR l_i = 1 TO l_len4-l_len
            LET l_flds[i].section4=l_flds[i].section4,' '
         END FOR
      END IF
      LET l_len = FGL_WIDTH(l_flds[i].section5)
      IF l_len5 > l_len AND l_len > 0 THEN
         FOR l_i = 1 TO l_len5-l_len
            LET l_flds[i].section5=l_flds[i].section5,' '
         END FOR
      END IF
      LET l_flds[i].section1 = cl_replace_str(l_flds[i].section1,"  ","　")
      LET l_flds[i].section2 = cl_replace_str(l_flds[i].section2,"  ","　")
      LET l_flds[i].section3 = cl_replace_str(l_flds[i].section3,"  ","　")
      LET l_flds[i].section4 = cl_replace_str(l_flds[i].section4,"  ","　")
      LET l_flds[i].section5 = cl_replace_str(l_flds[i].section5,"  ","　")
   END FOR
  
   IF l_flag1 = 'Y' THEN
      LET p_info1 = l_flds[1].section1,'　　　',l_flds[1].section2,'　　　',l_flds[1].section3,'　　　',l_flds[1].section4,'　　　',l_flds[1].section5
   END IF
   IF l_flag2 = 'Y' THEN
      LET p_info2 = l_flds[2].section1,'　　　',l_flds[2].section2,'　　　',l_flds[2].section3,'　　　',l_flds[2].section4,'　　　',l_flds[2].section5
   END IF
   IF l_flag3 = 'Y' THEN
      LET p_info3 = l_flds[3].section1,'　　　',l_flds[3].section2,'　　　',l_flds[3].section3,'　　　',l_flds[3].section4,'　　　',l_flds[3].section5
   END IF
   IF l_flag4 = 'Y' THEN
      LET p_info4 = l_flds[4].section1,'　　　',l_flds[4].section2,'　　　',l_flds[4].section3,'　　　',l_flds[4].section4,'　　　',l_flds[4].section5
   END IF
   IF l_flag5 = 'Y' THEN
      LET p_info5 = l_flds[5].section1,'　　　',l_flds[5].section2,'　　　',l_flds[5].section3,'　　　',l_flds[5].section4,'　　　',l_flds[5].section5
   END IF

   RETURN p_info1,p_info2,p_info3,p_info4,p_info5
END FUNCTION

################################################################################
# Descriptions...: #檢查 SQL 筆數
# Memo...........:
# Usage..........: cl_xg_rowcnt(p_table,p_where)
#                  RETURNING 回传参数
# Input parameter: p_table   
#                : p_where   
# Return code....: l_cnt 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_rowcnt(p_table,p_where)
   DEFINE p_table  LIKE type_t.chr1000
   DEFINE p_where  LIKE type_t.chr1000
   DEFINE l_sql    STRING
   DEFINE l_cnt    LIKE type_t.num10

   LET l_cnt = 0
   IF cl_null(p_where) THEN LET p_where = ' 1=1' END IF
#   LET l_sql = "SELECT COUNT(*) FROM ",p_table," WHERE ",p_where   #170525-00105#1 mark
   LET l_sql = "SELECT COUNT(1) FROM ",p_table," WHERE ",p_where    #170525-00105#1 add
   PREPARE cl_xg_rowcnt_pr FROM l_sql
   EXECUTE cl_xg_rowcnt_pr INTO l_cnt

   IF l_cnt <= 0 THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'lib-00073'
      LET g_errparam.extend = '!'
      LET g_errparam.popup = TRUE
      CALL cl_err()
      #190223-00010#2 add(s)
      IF g_bgjob = "Y" THEN
         CALL cl_gr_nodata_sendmail()
      END IF
      #190223-00010#2 add(e)
   END IF

   RETURN l_cnt
END FUNCTION

################################################################################
# Descriptions...: 
# Memo...........:
# Usage..........: CALL cl_xg_chk_column (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_gzgf000   
#                : p_sql   
#                : p_table 
#                : p_where 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_chk_column(p_gzgf000,p_sql,p_table,p_where)
   DEFINE p_gzgf000  LIKE gzgf_t.gzgf000
   DEFINE l_sql    STRING
   DEFINE l_cnt    LIKE type_t.num10
   DEFINE p_sql    STRING
   DEFINE p_table  LIKE type_t.chr1000
   DEFINE p_where  STRING
   DEFINE l_gzgg001  LIKE gzgg_t.gzgg001    #報表欄位代碼
   #DEFINE l_itgs004  LIKE itgs_t.itgs004   #gds04 欄位說明

   WHENEVER ERROR CONTINUE
    
   LET l_sql = "SELECT gzgg001 FROM gzgg_t ",  
#               " WHERE gzgg000 = ",p_gzgf000,        #160330-00019#1 mark
               " WHERE gzgg000 = '",p_gzgf000,"' ",   #160330-00019#1 add               
               "   AND gzgg010 = 1 "
   PREPARE cl_xg_chk_column_gzgg_pr FROM l_sql
   DECLARE cl_xg_chk_column_gzgg_cs CURSOR FOR cl_xg_chk_column_gzgg_pr
    
   FOREACH cl_xg_chk_column_gzgg_cs INTO l_gzgg001
      IF cl_null(p_sql) THEN
         LET l_sql = "SELECT COUNT(",l_gzgg001,") FROM ",g_rep_db CLIPPED,p_table CLIPPED
         PREPARE cl_xg_chk_column_sql1_pr FROM l_sql
         EXECUTE cl_xg_chk_column_sql1_pr INTO l_cnt
         IF STATUS THEN
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = 'azz-00148'
            LET g_errparam.extend = l_gzgg001
            LET g_errparam.popup = TRUE
            CALL cl_err()
            
            CALL cl_xg_view_addlog(l_gzgg001||cl_getmsg('azz-00148',g_rlang))
#            LET g_success = 'N'        #160330-00019#1 mark
            LET g_rep_success = 'N'     #160330-00019#1 add
            EXIT FOREACH
         END IF
      ELSE
         LET l_sql = "SELECT COUNT(",l_gzgg001,") FROM ",p_table CLIPPED,
                     " WHERE ",p_where
         PREPARE cl_xg_chk_column_sql2_pr FROM l_sql
         EXECUTE cl_xg_chk_column_sql2_pr INTO l_cnt
         IF STATUS THEN
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = 'azz-00148'
            LET g_errparam.extend = l_gzgg001
            LET g_errparam.popup = TRUE
            CALL cl_err()

            CALL cl_xg_view_addlog(l_gzgg001||cl_getmsg('azz-00148',g_rlang))
#            LET g_success = 'N'        #160330-00019#1 mark
            LET g_rep_success = 'N'     #160330-00019#1 add
            EXIT FOREACH
         END IF
      END IF
   END FOREACH
END FUNCTION

################################################################################
# Descriptions...: 
# Memo...........:
# Usage..........: CALL cl_xg_view_bgjob (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_url   
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_view_bgjob(p_url)
   DEFINE p_url     STRING
   DEFINE req        com.HTTPRequest
   DEFINE resp       com.HTTPResponse
   DEFINE l_str      STRING

   TRY 
      LET req = com.HTTPRequest.Create(p_url)
      CALL req.doRequest()
      LET resp = req.getResponse()
      IF resp.getStatusCode() != 200 THEN
         DISPLAY "HTTP Error ("||resp.getStatusCode()||") ",resp.getStatusDescription()
      ELSE
         LET l_str = resp.getTextResponse()
         DISPLAY "HTTPResponse : OK!"
         DISPLAY l_str
      END IF
   CATCH
      DISPLAY "ERROR :",STATUS||" ("||SQLCA.SQLERRM||")" 
   END TRY
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_xg_view_logfile() (传入参数)
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_view_logfile()
   DEFINE l_file  STRING,
          l_s     STRING,
          l_cmd   STRING,
          i       SMALLINT
   DEFINE p_interval     INTERVAL SECOND TO FRACTION(5)
   DEFINE l_filename STRING   #170210-00028#1 add
   DEFINE l_time  STRING   #190503-00009#2 add   #紀錄LOG開始時間

   #170828-00025#1 add(s)
   #將錯誤訊息DISPLAY出來，寫入SCHEDULE的log中
   IF g_bgjob = "Y" THEN
      DISPLAY "g_prog:",g_prog," g_rep_code:",g_rep_code," ",g_err_msg
   END IF
   #170828-00025#1 add(e)

   #記錄此次呼叫的 method name
#   LET l_file = "xtragrid-", TODAY USING 'YYYYMMDD', ".log"   #161117-00037#1 mark
#   LET l_file = "xtragrid-",g_ldate USING 'YYYYMMDD',".log"    #161117-00037#1 add   #170210-00028#1 mark
   #170210-00028#1 add(s)
   LET l_filename = "xtragrid-",g_ldate USING 'YYYYMMDD',".log"
   LET l_file = os.Path.join(FGL_GETENV("TEMPDIR"),l_filename)
   #170210-00028#1 add(e)   
   LET channel = base.Channel.create()

   CALL channel.openFile(l_file,  "a")
   IF STATUS = 0 THEN
      CALL channel.setDelimiter("")
      #紀錄傳入的 XML 字串
      #LET l_s = "#--------------------------- (", CURRENT YEAR TO SECOND, ") ----------------------------#"
#      LET l_s = "#--------------------------- (", cl_get_current(), ") ----------------------------#"   #190503-00009#2 mark   
      #190503-00009#2 add(s)
      LET l_time = cl_get_current()
      LET l_s = "#--------------------------- (", l_time, ") ----------------------------#"
      ##190503-00009#2 add(e)
      CALL channel.write(l_s)
      #190503-00009#2 add(s)
      #先在LOG開始寫入START的關鍵字
      IF cl_null(g_schedule_id) THEN
         LET l_s = "START report_sn:",g_msgparam.rpt_code
         CALL channel.write(l_s)
      ELSE
         LET l_s = "START report_sn:",g_msgparam.rpt_code,", g_schedule_id:",g_schedule_id,", g_schedule_seq:",g_schedule_seq
         CALL channel.write(l_s)
      END IF
      CALL channel.write(l_time)
      #190503-00009#2 add(e)
      LET l_s = "Program:",g_prog, " , User:",g_user," , Enterprise:",g_enterprise
      LET l_s = l_s," , Account:",g_account   #150323 add
      LET l_s = l_s, ", URL:",g_url           #151216-00001#1 add
      CALL channel.write(l_s)
      IF NOT cl_null(g_schedule_id) THEN
         LET l_s = "指定時間於背景執行 g_bgjob:",g_bgjob," ,g_schedule_id:",g_schedule_id," ,g_schedule_seq:",g_schedule_seq," ,g_msgparam.rpt_code:",g_msgparam.rpt_code          
         CALL channel.write(l_s)
         LET l_s = "g_run_cmd:",g_run_cmd
         CALL channel.write(l_s)
         LET l_s = "g_run_result:",g_run_result, " ,g_err_msg:",g_err_msg
         CALL channel.write(l_s)         
      ELSE
         IF g_bgjob = 'Y' THEN
            LET l_s = "立即於背景執行 g_bgjob:",g_bgjob," ,g_schedule_id:",g_schedule_id," ,g_schedule_seq:",g_schedule_seq," ,g_msgparam.rpt_code:",g_msgparam.rpt_code                   
            CALL channel.write(l_s)
            LET l_s = "g_run_cmd:",g_run_cmd
            CALL channel.write(l_s)
            LET l_s = "g_run_result:",g_run_result, " ,g_err_msg:",g_err_msg
            CALL channel.write(l_s)   
            #190107-00075#8 add(s)
            IF g_schedule_d.getLength() > 0 THEN
               LET l_s = "收件者資訊"
               CALL channel.write(l_s)
               FOR i = 1 TO g_schedule_d.getLength()
                  LET l_s = "   g_schedule_d[",i,"].gzpe007(形式):",g_schedule_d[i].gzpe007
                  CALL channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe004(員工編號):",g_schedule_d[i].gzpe004
                  CALL channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe009(Email address):",g_schedule_d[i].gzpe009
                  CALL channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe013(進階設定):",g_schedule_d[i].gzpe013
                  CALL channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe010(檔案格式):",g_schedule_d[i].gzpe010
                  CALL channel.write(l_s)               
               END FOR
            END IF
            #190107-00075#8 add(e)            
         ELSE
            #150804 add(s)
            LET l_s = "立即於前景執行 "
            CALL channel.write(l_s)            
            #150804 add(e)
            CALL channel.write(g_err_msg)   #紀錄樣板資訊 #190219-00005#1 add
         END IF
      END IF

    

      IF g_errlog.getlength() = 0 THEN #有錯誤
         CALL channel.write("Successful !")
         #190503-00009#2 add(s)
         LET l_s = "END"
         CALL channel.write(l_s)
         #190503-00009#2 add(e)
         LET l_s = "#------------------------------------------------------------------------------#"
         CALL channel.write(l_s)
      ELSE
         CALL channel.write("Error : ")
         FOR i = 1 TO g_errlog.getlength()
            CALL channel.write(g_errlog[i])
         END FOR
         #190503-00009#2 add(s)
         LET l_s = "END"
         CALL channel.write(l_s)
         #190503-00009#2 add(e)
         LET l_s = "#------------------------------------------------------------------------------#"
         CALL channel.write(l_s)
      END IF

      CALL channel.close()

      IF os.Path.chrwx(l_file CLIPPED,438) THEN END IF 
   END IF

END FUNCTION

################################################################################
# Descriptions...: 將錯誤訊息寫入陣列
# Memo...........:
# Usage..........: CALL cl_xg_view_addlog (p_str)
#                  RETURNING 回传参数
# Input parameter: p_str   
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_view_addlog(p_str)
   DEFINE p_str   STRING
   DEFINE l_n     LIKE type_t.num5

   LET l_n = g_errlog.getlength()
   LET g_errlog[l_n+1] = p_str

END FUNCTION

################################################################################
# Descriptions...: 取得匯出檔名
# Memo...........:
# Usage..........: CALL cl_xg_view_export_filename (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_gzgf000   
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_view_export_filename(p_gzgf000)
   DEFINE p_gzgf000     LIKE gzgf_t.gzgf000
   #DEFINE l_itcw        RECORD LIKE itcw_t.* #gcw_file CR報表檔名設定檔
   DEFINE l_gzgf        RECORD LIKE gzgf_t.*
   DEFINE p_code        LIKE type_t.chr1
   DEFINE l_filename    STRING
   DEFINE l_min,l_max   LIKE type_t.num5
   DEFINE l_str         STRING
   DEFINE l_item_name   STRING
   DEFINE l_zx02        LIKE type_t.chr80 #zx_file.zx02    #131203先不用
   #DEFINE l_zx04        LIKE zx_file.zx04    #131203先不用
   #DEFINE l_gaz06       LIKE gaz_file.gaz06 #04010 mark
   DEFINE l_gzzal003    LIKE gzzal_t.gzzal003 #gaz_file.gaz06
   #DEFINE l_gaz03       LIKE gaz_file.gaz03  #131203先不用
   DEFINE l_time        LIKE type_t.chr100

   #檢查是否有自訂命名，若無則依標準命名
   SELECT * INTO l_gzgf.* FROM gzgf_t
    WHERE gzgf000 = p_gzgf000
#       INITIALIZE l_itcw.* TO NULL
#       SELECT * INTO l_itcw.* FROM itcw_t
#        WHERE gcw01=l_gzgg.gzgf001
#          AND gcw02=l_gzgg.gzgf002
#          AND gcw03=l_gzgg.gzgf005
#          AND gcw04=l_gzgg.gzgf004

   #有設定自訂檔名(gcw) 報表檔名第一段~第六段
#   IF (NOT cl_null(l_itcw.itcw005)) OR (NOT cl_null(l_itcw.itcw006)) OR (NOT cl_null(l_itcw.itcw007)) OR (NOT cl_null(l_itcw.itcw008)) OR
#      (NOT cl_null(l_itcw.itcw009)) OR (NOT cl_null(l_itcw.itcw010)) THEN
          SELECT gzzal003 INTO l_gzzal003 FROM gzzal_t WHERE gzzal001=g_prog AND gzzal002 =g_rlang  #131203先暫用
          LET l_time=TIME
          LET p_code = l_gzgf.gzgf001
          #LET l_item_name = g_prog CLIPPED,"|",l_gzzal003 CLIPPED,"|",g_plant CLIPPED,"|",g_dbs CLIPPED,"|",g_user CLIPPED,"|",l_zx02 CLIPPED,"|",g_pdate CLIPPED,"|", l_time CLIPPED   #141016 mark
          LET l_item_name = g_prog CLIPPED,"|",l_gzzal003 CLIPPED,"|",g_site CLIPPED,"|",g_dbs CLIPPED,"|",g_user CLIPPED,"|",l_zx02 CLIPPED,"|",g_pdate CLIPPED,"|", l_time CLIPPED   #141016 add         
#          CALL cl_xg_filename(l_item_name,l_itcw.itcw005) RETURNING l_str  #CR報表檔名設定
          LET l_filename = l_filename CLIPPED,l_str CLIPPED
#          CALL cl_xg_filename(l_item_name,l_itcw.itcw006) RETURNING l_str  #CR報表檔名設定
          LET l_filename = l_filename CLIPPED,l_str CLIPPED
#          CALL cl_xg_filename(l_item_name,l_itcw.itcw007) RETURNING l_str  #CR報表檔名設定
          LET l_filename = l_filename CLIPPED,l_str CLIPPED
#          CALL cl_xg_filename(l_item_name,l_itcw.itcw008) RETURNING l_str  #CR報表檔名設定
          LET l_filename = l_filename CLIPPED,l_str CLIPPED
#          CALL cl_xg_filename(l_item_name,l_itcw.itcw009) RETURNING l_str  #CR報表檔名設定
          LET l_filename = l_filename CLIPPED,l_str CLIPPED
#          CALL cl_xg_filename(l_item_name,l_itcw.itcw010) RETURNING l_str  #CR報表檔名設定
          LET l_filename = l_filename CLIPPED,l_str CLIPPED
          #刪除最後一個分隔符號
          LET l_str = l_filename CLIPPED
          LET l_str = l_str.substring(1,l_str.getLength()-1) CLIPPED
          LET l_filename = l_str CLIPPED
#       ELSE  #不自訂命名，依標準命名
          LET l_filename = g_site, "_", g_prog CLIPPED, "_", g_user   #組合檔名
#       END IF
 
   RETURN l_filename
END FUNCTION

################################################################################
# Descriptions...: 取得預設群組要
# Memo...........:
# Usage..........: CALL cl_xg_view_gzgf014 (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_gzgf000   樣板id
#                : 
# Return code....: l_grups   
#                : 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_view_gzgf014(p_gzgf000)
   DEFINE p_gzgf000  LIKE gzgf_t.gzgf000
   DEFINE l_grups    LIKE gzgf_t.gzgf014  
   DEFINE l_n        LIKE type_t.num5
   DEFINE l_gzgg001  LIKE gzgg_t.gzgg001
   DEFINE l_gzgf014  LIKE gzgf_t.gzgf014

   LET l_grups = NULL   #先不用排序 131203-(S)
   #DECLARE p_xg_view_itgs036_cs CURSOR FOR 
    #SELECT itgs003,itgs036 FROM itgs_t
     #WHERE itgs000 = p_gzgg000
       #AND itgs036 IS NOT NULL
     #ORDER BY itgs036
   #LET l_n = 1
   #FOREACH p_xg_view_itgs036_cs INTO l_itgs003,l_itgs036
      #IF cl_null(l_itgs003) THEN CONTINUE FOREACH END IF
      #IF l_n = 1 THEN
         #LET l_grups = l_itgs003
      #ELSE
         #LET l_grups = l_grups CLIPPED,",",l_itgs003 CLIPPED 
      #END IF
      #LET l_n = l_n + 1
   #END FOREACH
   
   IF cl_null(l_grups) THEN LET l_grups = " " END IF

   RETURN l_grups
END FUNCTION

################################################################################
# Descriptions...:  取出TEMP TABLE中的VARCHAR欄位，並將值為NULL的部份UPDATE成一個空白
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_table   
#                :    
# Return code....: 
#                : 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_view_updatenull(p_table)
   DEFINE p_table  LIKE type_t.chr100
   DEFINE l_sql    STRING
   DEFINE l_column LIKE type_t.chr100
   DEFINE l_type   LIKE type_t.chr100

   LET l_sql = "SELECT COLUMN_NAME FROM ALL_TAB_COLUMNS ",
               " WHERE TABLE_NAME='",UPSHIFT(p_table),"'",
               "   AND DATA_TYPE='VARCHAR2'"
   PREPARE cl_xg_view_updatenull_pr FROM l_sql
   DECLARE cl_xg_view_updatenull_cs CURSOR FOR cl_xg_view_updatenull_pr
   FOREACH cl_xg_view_updatenull_cs INTO l_column,l_type
      LET l_sql = "UPDATE ",g_rep_db CLIPPED,p_table CLIPPED,
                  "   SET ",l_column," = ' '",
                  " WHERE ",l_column," IS NULL"
      PREPARE cl_xg_view_updatenull_pr2 FROM l_sql
      EXECUTE cl_xg_view_updatenull_pr2
   END FOREACH
END FUNCTION

################################################################################
# Descriptions...: drop暫存檔
# Memo...........:
# Usage..........: CALL cl_xg_drop_tmptable (传入参数)
#                  RETURNING 回传参数
# Input parameter: 
#                : 
# Return code....: 
#                : 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_drop_tmptable()
   DEFINE l_sql STRING
   DEFINE i     SMALLINT

   #LET g_rep_db = "dsrept."  #140910 mark
   #LET g_report = g_rep_db CLIPPED
   
   FOR i = 1 TO g_rep_tmpname.getlength()
      IF NOT cl_null(g_rep_tmpname[i]) THEN
         #LET l_sql = "DROP TABLE ",g_report CLIPPED,g_rep_tmpname[i]
         LET l_sql = "DROP TABLE ",g_rep_db CLIPPED,g_rep_tmpname[i] #ODI-ORA#
#         DISPLAY l_sql
         EXECUTE IMMEDIATE l_sql
      END IF
   END FOR
END FUNCTION

################################################################################
# Descriptions...: create 暫存檔及組insert子句cl_xg_create_tmptable
# Memo...........:
# Usage..........: CALL cl_xg_create_tmptable(p_sql,p_cnt)
#                  RETURNING 回传参数
# Input parameter: p_sql
#                : p_cnt
# Return code....: BOOLEAN  #g_rep_tmpname  暫存檔陣列
#                :          #g_rep_ins_prep insert子句陣列
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_create_tmptable(p_sql,p_cnt)
   DEFINE p_sql          STRING
   DEFINE p_cnt          LIKE type_t.num5  
   DEFINE l_path         STRING
   DEFINE l_doc          om.DomDocument
   DEFINE l_rootnode     om.DomNode
   DEFINE l_nodes        om.NodeList
   DEFINE l_curnode      om.DomNode
   DEFINE l_nodes2       om.NodeList
   DEFINE l_curnode2     om.DomNode
   DEFINE l_tagname      STRING
   DEFINE l_tagname2     STRING
   DEFINE l_colname      VARCHAR(100)
   DEFINE l_ref_table    VARCHAR(100)
   DEFINE l_ref_col      VARCHAR(100)
   DEFINE i,j,l_cols,l_cnt   SMALLINT
   DEFINE l_table        STRING  
   DEFINE l_sql          STRING
   DEFINE l_success      VARCHAR(1)
   DEFINE l_sql1         STRING  #將insert prep併進來   
   #DEFINE p_type         LIKE dzgc_t.dzgc003                     #欄位處理類型
   DEFINE ls_record_tmp  STRING
   DEFINE ls_def_record  STRING
   DEFINE ls_dzgc004     LIKE dzgc_t.dzgc004  #變數代號
   DEFINE ls_dzgc006     LIKE dzgc_t.dzgc006  #自訂欄位
   DEFINE ls_dzgd004     LIKE dzgd_t.dzgd004  #參考欄位型態
   DEFINE ls_dzgd004_tmp LIKE dzgd_t.dzgd004  
   DEFINE ls_dzgd006     LIKE dzgd_t.dzgd006  #公式
   DEFINE ls_dzgd004_str STRING
   DEFINE ls_gztz001 	 LIKE gztz_t.gztz001 	
   DEFINE l_dzgc004_str  STRING
   
   DEFINE l_gzgf000      LIKE gzgf_t.gzgf000   #樣板ID
   DEFINE l_gzgg001      LIKE gzgg_t.gzgg001   #報表欄位代碼
   DEFINE l_gzgg026      LIKE gzgg_t.gzgg026   #欄位別名
   DEFINE ls_dzgc009     LIKE dzgc_t.dzgc009   #客製標示

   LET l_sql = p_sql
   LET l_cnt = p_cnt
   
   IF cl_null(l_sql) THEN
      DISPLAY "gzgg,dzgd及dzgc中無欄位資料，無法產生temp table"
      RETURN FALSE   #161117-00037#1 add 
   END IF

   LET g_rep_tmpname[l_cnt] = cl_xg_tmptable(g_prog,l_sql,p_cnt) CLIPPED   # 產生Temp Table
   #DISPLAY "l_table:",l_table
   LET l_success = 'Y'  
   IF g_rep_tmpname[l_cnt] = -1 THEN
      LET l_success = 'N'
      DISPLAY "Create failue! ls_def_record:",ls_def_record
   END IF
 
   LET l_cnt = l_cnt + 1

   IF l_success = 'Y' THEN
      RETURN TRUE
   ELSE
      RETURN FALSE
   END IF
END FUNCTION

################################################################################
# Descriptions...: 取AP Server的hostname (若不同ap server未依慣例而設定相同hostname時,將會造成系統誤判誤砍)
# Memo...........:
# Usage..........: cl_used_ap_hostname()
#                  RETURNING 回传参数
# Input parameter: 
#                : 
# Return code....: AP_Server_Hostname  String
#                : 
# Date & Author..: 140703 By 02716
# Modify.........: #151119-00011 改用cl_get_ap_hostname
################################################################################
PUBLIC FUNCTION cl_used_ap_hostname()
   DEFINE ls_host   STRING
   DEFINE ch base.Channel

   IF cl_null(gs_hostname) THEN
      LET ch = base.Channel.create()
      CALL ch.setDelimiter(".")
      CALL ch.openPipe("hostname","r")

      LET ls_host = ""
      WHILE ch.read(ls_host)
      END WHILE
   ELSE
      LET ls_host = gs_hostname
   END IF

   CALL ch.close()
   RETURN ls_host.trim()
END FUNCTION

################################################################################
# Descriptions...: 把字串轉為ASCII CODE
# Memo...........:
# Usage..........: CALL cl_xg_trans_ascii(l_str)
#                  RETURNING l_str_ascii
# Input parameter: l_str
# Return code....: l_str_ascii
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_trans_ascii(l_str)
   DEFINE l_str,l_str_ord     STRING,
          l_str_ascii         STRING,
          i                   LIKE type_t.num5

   LET l_str_ascii=""
   FOR i=1 TO l_str.getLength()
      LET l_str_ord = ORD(l_str.subString(i,i))
      IF i=1 THEN
         LET l_str_ascii = l_str_ord.trim()
      ELSE
         LET l_str_ascii = l_str_ascii.trim() ,'|',l_str_ord.trim()
      END IF
   END FOR
   RETURN l_str_ascii
END FUNCTION

################################################################################
# Descriptions...: 報表檔名設定
# Memo...........:
# Usage..........: CALL cl_xg_filename (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_item_name   (報表檔名全部選項，以"|"隔開)
#                : p_this(選擇項目)
# Return code....: 檔名段落
#                : 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_filename(p_item_name,p_this)
   DEFINE p_item_name       STRING               #報表檔名全部選項，以"|"隔開
   DEFINE p_this            STRING  #選擇項目
   #1: zaw01 程式代號
   #2: gaz03 程式名稱(須依照g_rlang來取得資料)
   #3: 營運中心編號
   #4: 營運中心名稱
   #5: zaw05使用者代號(zx01)
   #6: zx02使用者名稱(zx02)
   #7: 2008/12/17年月日 10碼
   #8: 05:38:59時分秒 8碼
   #A~T: p1~p20
   DEFINE l_item            DYNAMIC ARRAY OF RECORD
                               val     STRING
                            END RECORD
   DEFINE l_tok             base.StringTokenizer
   DEFINE l_i               INTEGER

   LET l_tok = base.StringTokenizer.createExt(p_item_name CLIPPED,"|","",TRUE)  #指定分隔符號
   WHILE l_tok.hasMoreTokens()  #依序取得子字串
      LET l_i = l_i+1
      LET l_item[l_i].val = l_tok.nextToken()
      #刪除windows不支援的檔名特殊符號
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '\\', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '/', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, ':', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '*', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '?', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '"', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '<', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '>', '')
      LET l_item[l_i].val = cl_replace_str(l_item[l_i].val, '|', '')
      LET l_item[l_i].val = l_item[l_i].val,"_"   #加上分隔符號
   END WHILE

   CASE p_this
      WHEN "1"
         RETURN l_item[1].val
      WHEN "2"
         RETURN l_item[2].val
      WHEN "3"
         RETURN l_item[3].val
      WHEN "4"
         RETURN l_item[4].val
      WHEN "5"
         RETURN l_item[5].val
      WHEN "6"
         RETURN l_item[6].val
      WHEN "7"
         RETURN l_item[7].val
      WHEN "8"
         RETURN l_item[8].val
      WHEN "A"
         RETURN l_item[9].val
      WHEN "B"
         RETURN l_item[10].val
      WHEN "C"
         RETURN l_item[11].val
      WHEN "D"
         RETURN l_item[12].val
      WHEN "E"
         RETURN l_item[13].val
      WHEN "F"
         RETURN l_item[14].val
      WHEN "G"
         RETURN l_item[15].val
      WHEN "H"
         RETURN l_item[16].val
      WHEN "I"
         RETURN l_item[17].val
      WHEN "J"
         RETURN l_item[18].val
      WHEN "K"
         RETURN l_item[19].val
      WHEN "L"
         RETURN l_item[20].val
      WHEN "M"
         RETURN l_item[21].val
      WHEN "N"
         RETURN l_item[22].val
      WHEN "O"
         RETURN l_item[23].val
      WHEN "P"
         RETURN l_item[24].val
      WHEN "Q"
         RETURN l_item[25].val
      WHEN "R"
         RETURN l_item[26].val
      WHEN "S"
         RETURN l_item[27].val
      WHEN "T"
         RETURN l_item[28].val
      OTHERWISE
         RETURN ""
   END CASE
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_xg_tmptable (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_prog   
#                : p_sql 
#                : p_cnt  
# Return code....: 傳回自動產生的 Temp Table Name
#                : 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_tmptable(p_prog,p_sql,p_cnt)
   DEFINE p_prog                 VARCHAR(100),
          p_sql,l_length         STRING,
          p_cnt                  DEC(5,0),           
          l_table,l_str          STRING,
          l_name                 STRING,
          l_time                 STRING,               
          l_n                    DEC(10,0),
          l_tok_column           base.StringTokenizer,
          l_tok_table            base.StringTokenizer,
          l_table_name           VARCHAR(100),
          l_field_name           VARCHAR(100),
          l_alias_name           VARCHAR(100),
          l_datatype             VARCHAR(100),
          l_cnt_dot,l_cnt_comma  DEC(5,0),
          l_p                    DEC(5,0),
          l_p2                   DEC(5,0),
          l_sql                  STRING,
          l_db_type              STRING,
          l_mssqlarea            STRING,                
          l_dbname               STRING                 
   DEFINE l_str1                 STRING                
   DEFINE l_hostname             STRING             #Host Name只取英數字
   DEFINE l_tbname_length        DEC(5,0)           #Table Name 長度限制 
   DEFINE g_sql                  STRING
   DEFINE l_repcnt_str           STRING

   #LET g_rep_db='dsrept.' #140910 mark
   IF g_rep_db IS NULL OR g_rep_db = '' THEN LET g_rep_db = 'dsrept.' END IF  #140910 add
   #LET g_report = g_rep_db CLIPPED,"." 

   LET g_rep_print_cnt = g_rep_print_cnt + 1 
#   DISPLAY "g_rep_print_cnt:",g_rep_print_cnt 

   LET l_tbname_length = 30   #Table Name 長度限制

   LET l_str = FGL_GETPID()       #使用GETPID抓PID的值
   LET l_str = l_str.trim()
   #LET l_table = p_prog CLIPPED, l_str CLIPPED
   IF g_rep_print_cnt < 10 THEN
      LET l_repcnt_str = g_rep_print_cnt CLIPPED
      LET l_repcnt_str = "0",l_repcnt_str 
   ELSE
      LET l_repcnt_str = g_rep_print_cnt CLIPPED
   END IF 
   DISPLAY "l_repcnt_str:",l_repcnt_str

   LET g_rep_tmpname[p_cnt] = g_rep_code CLIPPED, p_cnt USING '&',l_repcnt_str CLIPPED, l_str CLIPPED
#   DISPLAY "XG temp table name:",g_rep_tmpname[p_cnt]          #170828-00025#1 mark
   DISPLAY "XG temp table name: dsrept.",g_rep_tmpname[p_cnt]   #170828-00025#1 add
   ##Table Name 長度限制

   IF g_rep_tmpname[p_cnt].getLength() > l_tbname_length THEN      
      LET g_rep_tmpname[p_cnt] = g_rep_tmpname[p_cnt].subString(1,l_tbname_length)   #prog+PID
   ELSE
      LET l_str1 = l_hostname CLIPPED
      ##Table Name加上Host Name後若超過長度限制,則取局部Host Name       
      IF (NOT cl_null(l_hostname)) AND (g_rep_tmpname[p_cnt].getLength()+l_hostname) > l_tbname_length THEN
         LET l_n = l_tbname_length - g_rep_tmpname[p_cnt].getLength()
         IF l_hostname MATCHES "*[a-z]*" THEN   #not IP,取hostname前面
            LET l_str1 = l_hostname.subString(1,l_n)
         ELSE                                   #IP,取hostname後面
            LET l_n = l_hostname.getLength() - l_n + 1
            LET l_str1 = l_hostname.subString(l_n,l_hostname.getLength())
         END IF
      END IF
      LET g_rep_tmpname[p_cnt] = g_rep_code CLIPPED, p_cnt USING '&', l_repcnt_str CLIPPED, l_str CLIPPED, l_str1 CLIPPED   #prog+PID+hostname
   END IF    
    
   IF g_report IS NULL OR g_report = '' THEN LET g_report = 'dsrept' END IF

   #CALL cl_ins_del_sid(2, g_plant) #FUN-980030:將原本的SESSIONID刪除

#   CALL cl_db_connect(g_report,FALSE)  #140910 mark   #151216-00001#1 mark
   CALL cl_db_connect(g_report,FALSE)  #140910 mark    #151216-00001#1 add
   #CALL cl_db_connect(g_rep_db,FALSE)   #140910 add 
   
   SET CONNECTION g_report   #140910 add
   #SET CONNECTION "dsrept"   #140910 mark
   #SET CONNECTION g_rep_db    #140910 add     

   LET g_sql = "DROP TABLE ", g_rep_tmpname[p_cnt] CLIPPED  #ODI-ORA#
   DISPLAY "drop table g_sql:",g_sql 
   PREPARE drop_prep FROM g_sql
   
   LET l_tok_table = base.StringTokenizer.create(p_sql,".")
   LET l_cnt_dot = l_tok_table.countTokens()
   LET l_tok_table = base.StringTokenizer.create(p_sql,",")
   LET l_cnt_comma = l_tok_table.countTokens()
   #IF cl_null(l_dbname) THEN LET l_dbname = 'dsdemo' END IF
   IF cl_null(l_dbname) THEN LET l_dbname = g_dbs END IF    

   WHILE l_tok_table.hasMoreTokens()
      #DISPLAY l_tok_table.nextToken()
      LET l_name = l_tok_table.nextToken()
      LET l_p = l_name.getIndexOf(".",1)
      LET l_p2 = l_name.getIndexOf(".",l_p+1)
      LET l_alias_name = l_name.subString(1,l_p-1)
      LET l_table_name = l_name.subString(l_p+1,l_p2-1)
      LET l_field_name = l_name.subString(l_p2+1,l_name.getLength())
       
#      CALL cl_xg_get_column_info(l_dbname,l_table_name,l_field_name)   #150901-00021#1 mark
      CALL cl_rpt_get_column_info(l_dbname,l_table_name,l_field_name)   #150901-00021#1 add
         RETURNING l_datatype,l_length       
      
      IF NOT cl_null(l_sql) THEN
         LET l_sql = l_sql,","
      END IF
      IF l_datatype = "smallint" OR l_datatype = "integer" OR l_datatype = "date"
         OR l_datatype = "datetime" 
         OR l_datatype = "blob"  OR l_datatype = "byte"
         OR l_datatype = "int4"  OR l_datatype = "int2" #240612-00039#3 add
         #PGS(D)-02667
         #PGS(D)-02667
      THEN
         LET l_sql = l_sql, l_alias_name CLIPPED," ",l_datatype
      ELSE
         LET l_sql = l_sql, l_alias_name CLIPPED," ",l_datatype,
                     "(",l_length,")"
      END IF
   END WHILE
   LET l_sql = "(",l_sql,")"
    
   LET g_sql = "CREATE TABLE ",g_rep_tmpname[p_cnt] CLIPPED,l_sql CLIPPED
   #PGS(D)-11084(s)
   #PGS(D)-11084(e)
   DISPLAY "Create sql:",g_sql    
   PREPARE create_prep FROM g_sql
   IF SQLCA.sqlcode THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = SQLCA.sqlcode
      LET g_errparam.extend = 'create_prep:'
      LET g_errparam.popup = TRUE
      CALL cl_err()

      DISPLAY "Create TABLE Failure!"
      DISPLAY "p_sql:",p_sql
      RETURN -1
   END IF

   EXECUTE create_prep
   #141216 dsrept資料庫空間不足 add (s)
   IF SQLCA.sqlcode = '-6372' THEN
      #161117-00037#1 add(s)
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'lib-00420'
      LET g_errparam.extend = ''
      LET g_errparam.popup = TRUE
      CALL cl_err() 
      #161117-00037#1 add(e)
      DISPLAY "[ERROR] 報表暫存table使用的資料庫(dsrept)空間不足，請聯繫資料庫管理者(DBA)"      
      RETURN -1
   END IF
   #141216 dsrept資料庫空間不足 add (e)
   IF SQLCA.sqlcode THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = SQLCA.sqlcode
      LET g_errparam.extend = 'create_table:'
      LET g_errparam.popup = TRUE
      CALL cl_err()

      #DISPLAY "Create sql:",g_sql
      DISPLAY "Create TABLE Failure!"
      DISPLAY "p_sql:",p_sql
      RETURN -1
   END IF

   #181017-00035#2 add(s)
      #240612-00039#3 modify(s)
      IF g_db_branch="KINGBASE" THEN
         LET g_sql = "alter table ",g_rep_tmpname[p_cnt] CLIPPED," set unlogged "
         PREPARE cl_xg_unlogged_prep FROM g_sql
         EXECUTE cl_xg_unlogged_prep
      ELSE
         LET g_sql = "alter table ",g_rep_tmpname[p_cnt] CLIPPED," nologging "
         PREPARE cl_xg_nologging_prep FROM g_sql
         EXECUTE cl_xg_nologging_prep
      END IF
      #240612-00039#3 modify(e)
   #PGS(D)-12787 add -S
   #PGS(D)-12787 add -E

   #181017-00035#2 add(e)

   LET g_sql = "GRANT ALL ON ",g_rep_tmpname[p_cnt] CLIPPED," TO PUBLIC"
   PREPARE grant_prep FROM g_sql
   EXECUTE grant_prep

   CALL cl_db_disconnect_current()     
   SET CONNECTION g_dbs       

   #CALL cl_ins_del_sid(1, g_plant) #儲存新的SESSIONID  #T100先不用

   #LET g_temptable = l_table

   RETURN g_rep_tmpname[p_cnt]   # 傳回自動產生的 Temp Table Name 
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_xg_del_data (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_table
#                : 
# Return code....: 
#                : 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_del_data(p_table)
   DEFINE  p_table      STRING
   DEFINE  l_db_type    STRING
   DEFINE  g_sql        STRING
   #190426-00041#1 add(s)
   DEFINE  l_table      LIKE type_t.chr100      #PROCEDURE宣告是用Varchar，所以會將傳入的值改由此變數接
   DEFINE  l_cnt        LIKE type_t.num10       
   #190426-00041#1 add(e)

   #LET g_rep_db = "dsrept."  #140910 mark
   #LET g_report = g_rep_db CLIPPED,"."

   #180719-00037#1 mark(s)
#   LET g_sql = "DELETE FROM ",g_rep_db CLIPPED,p_table CLIPPED," WHERE 1=1"   
#   #LET g_sql = "DELETE FROM ",g_report CLIPPED,p_table CLIPPED," WHERE 1=1"
#   PREPARE deldata_prep FROM g_sql   #180719-00037#1 mark
#
#   EXECUTE deldata_prep   #180719-00037#1 mark
   #180719-00037#1 mark(e)
   #190426-00041#1 mark(s)
#   #180719-00037#1 add(s)
#   LET g_sql = "TRUNCATE TABLE ",g_rep_db CLIPPED,p_table CLIPPED
#   PREPARE truncate_prep FROM g_sql
#   EXECUTE truncate_prep
#   #180719-00037#1 add(e)
   #190426-00041#1 mark(e)
   #190426-00041#1 add(s)
   #PGS(D)-05292  
   SELECT COUNT(1) INTO l_cnt FROM DBA_SOURCE WHERE OWNER='DSREPT' AND NAME = 'TRUNCATETABLE'
   #PGS(D)-05292  

   #判斷PROCEDURE是否存在，若否走使用原先方式
   IF l_cnt > 0 THEN
      LET l_table = p_table CLIPPED
      #PGS(D)-05292  
      PREPARE truncate_pre FROM "begin dsrept.truncatetable(?); end;"
      #PGS(D)-05292  
      EXECUTE truncate_pre USING l_table IN
   ELSE
      LET g_sql = "TRUNCATE TABLE ",g_rep_db CLIPPED,p_table CLIPPED
      PREPARE truncate_prep FROM g_sql
      EXECUTE truncate_prep
   END IF
   #190426-00041#1 add(e)
               
   IF SQLCA.sqlcode THEN
      LET p_table = -1

      #LET l_db_type=cl_db_get_database_type()
      LET l_db_type="ORA"
      IF (l_db_type="ORA" AND SQLCA.sqlcode=-206) THEN
         #IF g_bgerr THEN
            #INITIALIZE g_errparam TO NULL
            #LET g_errparam.code = 'lib-00082'
            #LET g_errparam.extend = ''
            #LET g_errparam.popup = TRUE
            #CALL cl_err()
         #ELSE
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = 'lib-00082'
            LET g_errparam.extend = ''
            LET g_errparam.popup = TRUE
            CALL cl_err()

         #END IF
      ELSE
         #IF g_bgerr THEN
            #INITIALIZE g_errparam TO NULL
            #LET g_errparam.code = SQLCA.sqlcode
            #LET g_errparam.extend = 'deldata_prep:'
            #LET g_errparam.popup = TRUE
            #CALL cl_err()
         #ELSE
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = SQLCA.sqlcode
            LET g_errparam.extend = 'deldata_prep:'
            LET g_errparam.popup = TRUE
            CALL cl_err()
         #END IF
      END IF
      CALL cl_ap_exitprogram("1")
   END IF
END FUNCTION

################################################################################
# Descriptions...: 取得欄位資訊
# Memo...........:
# Usage..........: CALL cl_xg_get_column_info(p_dbname,p_tabname,p_colname)
#                  RETURNING r_datatype,r_length
# Input parameter: l_dbname   
#                : l_table_name table代號  
#                : l_field_name 欄位代號  
# Return code....: r_datatype   資料型態
#                : r_length     欄位長度
# Date & Author..: 2014/7/17 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_get_column_info
################################################################################
PUBLIC FUNCTION cl_xg_get_column_info(p_dbname,p_tabname,p_colname)
   DEFINE p_dbname    VARCHAR(20)
   DEFINE p_tabname   VARCHAR(20)
   DEFINE p_colname   VARCHAR(20)
   DEFINE r_datatype  VARCHAR(15)
   DEFINE r_length    STRING
   #150901-00021#1 mark(s)
#   DEFINE l_db_type   VARCHAR(3)
#   DEFINE l_sql       STRING
#   DEFINE l_str       STRING
#   DEFINE l_tmp       STRING
#   DEFINE l_dbname    STRING 
#   DEFINE l_sch03     VARCHAR(20)
#   DEFINE l_cnt       DEC(5,0)
#   DEFINE l_module    LIKE dzea_t.dzea003
#   DEFINE l_gztd008   LIKE gztd_t.gztd008   
   #150901-00021#1 mark(e)

   #150901-00021#1 mark(s)
#   LET l_db_type='ORA'
#   
#   LET p_tabname = p_tabname CLIPPED
#
#   SELECT dzea003 INTO l_module FROM ds.dzea_t WHERE dzea001 = p_tabname
#   IF l_module = "ADZ" OR l_module = "AZZ" THEN
#      LET p_dbname ="ds"
#   END IF
#
#   LET p_colname = p_colname CLIPPED
#   
#   LET l_sql= " SELECT gztd003,gztd008 ",
#              "  FROM ",p_dbname CLIPPED,".dzeb_t, ",p_dbname CLIPPED,".gztd_t ",
#              " WHERE dzeb001='",p_tabname CLIPPED,"' ",
#              "   AND dzeb002='",p_colname CLIPPED,"' ",
#              "   AND dzeb006 = gztd001 "
#
#   LET l_str = "SELECT COUNT(1) FROM (",l_sql,") tsed"
#   PREPARE cl_xg_get_column_info_cnt_pre FROM l_str
#   EXECUTE cl_xg_get_column_info_cnt_pre INTO l_cnt
#   FREE cl_xg_get_column_info_cnt_pre  
#   IF l_cnt = 0 THEN
#      LET l_sql= " SELECT gztd003,gztd008 ",
#                 "  FROM ds.dzeb_t, ds.gztd_t ",
#                 " WHERE dzeb001='",p_tabname CLIPPED,"' ",
#                 "   AND dzeb002='",p_colname CLIPPED,"' ",
#                 "   AND dzeb006 = gztd001 "      
#   END IF
#   
#   PREPARE cl_xg_get_column_info_p FROM l_sql
#   IF SQLCA.SQLCODE THEN
#      INITIALIZE g_errparam TO NULL
#      LET g_errparam.code =  SQLCA.SQLCODE
#      LET g_errparam.extend = "prepare column_info: "
#      LET g_errparam.popup = TRUE
#      CALL cl_err()
#
#      RETURN '',''
#   END IF
# 
#   EXECUTE cl_xg_get_column_info_p INTO r_datatype,l_gztd008 
#   LET r_length = l_gztd008                              
   #150901-00021#1 mark(e)

   CALL cl_rpt_get_column_info(p_dbname,p_tabname,p_colname) RETURNING r_datatype,r_length   #150901-00021#1 add
   RETURN r_datatype,r_length   
END FUNCTION

################################################################################
# Descriptions...: 取得資料表資料列數
# Memo...........:
# Usage..........: CALL cl_xg_data_rowcnt(p_table)
#                  RETURNING r_cnt
# Input parameter: p_table        table名稱
# Return code....: r_cnt          資料表資料列數
# Date & Author..: 2014/7/17 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_data_rowcnt(p_table)
   DEFINE p_table  STRING
   DEFINE l_sql    STRING
   DEFINE r_cnt    LIKE type_t.num10

#   LET l_sql = "SELECT COUNT(*) FROM ",g_rep_db CLIPPED,p_table CLIPPED   #170525-00105#1 mark
   LET l_sql = "SELECT COUNT(1) FROM ",g_rep_db CLIPPED,p_table CLIPPED    #170525-00105#1 add
   PREPARE cl_xg_data_rowcnt_pre FROM l_sql
   EXECUTE cl_xg_data_rowcnt_pre INTO r_cnt
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code =  SQLCA.SQLCODE
      LET g_errparam.extend = 'cl_xg_data_rowcnt_pre:'
      LET g_errparam.popup = TRUE
      CALL cl_err()
   ELSE
      IF r_cnt <= 0 THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'lib-00073'
         LET g_errparam.extend = '!'
         LET g_errparam.popup = TRUE
         CALL cl_err()
         #190223-00010#2 add(s)
         IF g_bgjob = "Y" THEN
            CALL cl_gr_nodata_sendmail()
         END IF
         #190223-00010#2 add(e)
      END IF
   END IF
   RETURN r_cnt
END FUNCTION

################################################################################
# Descriptions...: XG簽核字串
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2014/10/16 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_apr_str()
   DEFINE l_oojj005         LIKE oojj_t.oojj005 #簽核代號
   DEFINE l_sql,l_str       STRING  
   DEFINE r_oojj005_desc    LIKE type_t.chr500

   #簽核
   #至oojj_t取得簽核代號
   LET l_sql = "SELECT oojj005 FROM oojj_t WHERE oojjstus = 'Y' ",
               "AND oojjent = ? AND oojjsite =? AND oojj001 =? AND oojj002 = 'ALL' "
                        
#   DECLARE cl_xg_oojj005_cs CURSOR FROM l_sql
   PREPARE cl_xg_oojj005_pr FROM l_sql  #150114 add    
   
   #第一階段,取得符合的ent,site,g_rep_code ex:(99,DSCTC,ainr100_x01)    
   EXECUTE cl_xg_oojj005_pr USING g_enterprise,g_site,g_rep_code INTO l_oojj005 
#   FOREACH cl_xg_oojj005_cs USING g_enterprise,g_site,g_rep_code INTO l_oojj005     
#      IF SQLCA.SQLCODE THEN
#         INITIALIZE g_errparam TO NULL
#         LET g_errparam.code =  SQLCA.SQLCODE
#         LET g_errparam.extend = 'oojj005:'
#         LET g_errparam.popup = FALSE
#         CALL cl_err()
#      END IF
#   END FOREACH
      
   #第二階段,取得符合的ent,default,g_rep_code ex:(99,default,aint100_x01)
   IF cl_null(l_oojj005) THEN    
      EXECUTE cl_xg_oojj005_pr USING g_enterprise,"default",g_rep_code INTO l_oojj005       
#      FOREACH cl_xg_oojj005_cs USING g_enterprise,"default",g_rep_code INTO l_oojj005       
#         IF SQLCA.SQLCODE THEN
#            INITIALIZE g_errparam TO NULL
#            LET g_errparam.code =  SQLCA.SQLCODE
#            LET g_errparam.extend = 'oojj005:'
#            LET g_errparam.popup = FALSE
#            CALL cl_err()
#         END IF
#      END FOREACH
   END IF
      
   #第三階段,取得符合的ent,site,ALL ex:(99,DSCTC,ALL)
   IF cl_null(l_oojj005) THEN   
      EXECUTE cl_xg_oojj005_pr USING g_enterprise,g_site,"ALL" INTO l_oojj005        
#      FOREACH cl_xg_oojj005_cs USING g_enterprise,g_site,"ALL" INTO l_oojj005        
#         IF SQLCA.SQLCODE THEN
#            INITIALIZE g_errparam TO NULL
#            LET g_errparam.code =  SQLCA.SQLCODE
#            LET g_errparam.extend = 'oojj005:'
#            LET g_errparam.popup = FALSE
#            CALL cl_err()
#         END IF
#      END FOREACH   
   END IF   
   
   #第四階段,取得符合的ent,default,ALL ex:(99,default,ALL)
   IF cl_null(l_oojj005) THEN   
      EXECUTE cl_xg_oojj005_pr USING g_enterprise,"default","ALL" INTO l_oojj005        
#      FOREACH cl_xg_oojj005_cs USING g_enterprise,"default","ALL" INTO l_oojj005        
#         IF SQLCA.SQLCODE THEN
#            INITIALIZE g_errparam TO NULL
#            LET g_errparam.code =  SQLCA.SQLCODE
#            LET g_errparam.extend = 'oojj005:'
#            LET g_errparam.popup = FALSE
#            CALL cl_err()
#         END IF
#      END FOREACH   
   END IF

   IF NOT cl_null(l_oojj005) THEN      
      LET l_str = s_aooi900_get_oojk004_str(l_oojj005,g_dlang)
      #LET l_str = cl_replace_str(l_str, ",", "|")
      IF NOT cl_null(l_str) THEN    
         LET r_oojj005_desc = l_str
      END IF
   END IF

   RETURN r_oojj005_desc
END FUNCTION

################################################################################
# Descriptions...: 取得樣板名稱
# Memo...........:
# Usage..........: CALL cl_xg_get_gzgh011(p_gzgf006,p_gzgf000,p_gzgf004,p_gzgf005,p_gzgf002,p_gzgg020)
#                  RETURNING r_gzgh011
# Input parameter: p_gzgf006      樣板類型
#                : p_gzgf002      樣板名稱
#                : p_gzgf004      用戶
#                : p_gzgf005      角色
#                : p_gzgf002      樣板代號
#                : p_gzgg020      圖表類型
# Return code....: r_gzgh011
# Date & Author..: 2014/8/19 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_get_gzgh011(p_gzgf006,p_gzgf000,p_gzgf004,p_gzgf005,p_gzgf002,p_gzgg020)
   DEFINE p_gzgf006      LIKE gzgf_t.gzgf006 
   DEFINE p_gzgf000      LIKE gzgf_t.gzgf000 
   DEFINE p_gzgf004      LIKE gzgf_t.gzgf004 
   DEFINE p_gzgf005      LIKE gzgf_t.gzgf005
   DEFINE p_gzgf002      LIKE gzgf_t.gzgf002
   DEFINE p_gzgg020      LIKE gzgg_t.gzgg020   
   DEFINE r_gzgh011      STRING
   DEFINE l_gzgf006_desc LIKE gzcbl_t.gzcbl004 #樣板類型 
   DEFINE l_gzgf000_desc LIKE gzgdl_t.gzgdl002 #樣板名稱 ex:系統標準樣板
   DEFINE l_gzgf004_desc LIKE ooag_t.ooag011   #用戶說明
   DEFINE l_gzgf005_desc LIKE gzyal_t.gzyal003 #角色說明
   DEFINE l_cnt_type     INTEGER               #圖表類型

   LET l_gzgf006_desc = ""
   LET l_gzgf000_desc = ""
   LET l_gzgf004_desc = ""
   LET l_gzgf005_desc = ""
   
   ###樣板類型
   #SELECT gzcbl004 INTO l_gzgf006_desc FROM gzcbl_t 
   # WHERE gzcbl001 = 143 AND gzcbl002 = p_gzgf006 AND gzcbl003 = g_lang
    
   CASE p_gzgf006
      WHEN "1"	#明細清單
         LET l_gzgf006_desc = cl_getmsg('lib-00051',g_rlang) 
      WHEN "2"	#明細清單+單頭
         LET l_gzgf006_desc = cl_getmsg('lib-00052',g_rlang) 
      WHEN "3"	#明細清單+子報表
         LET l_gzgf006_desc = cl_getmsg('lib-00053',g_rlang) 
      WHEN "4"	#明細清單+單頭+子報表
         LET l_gzgf006_desc = cl_getmsg('lib-00054',g_rlang) 
      WHEN "5"	#交叉分析表
         LET l_gzgf006_desc = cl_getmsg('lib-00055',g_rlang) 
      WHEN "6"	#樹狀表
         LET l_gzgf006_desc = cl_getmsg('lib-00056',g_rlang)
      #180724-00009#2 mark(s)
#      #141016 add (s)
#      WHEN "7" #明細清單+圖形
#         LET l_gzgf006_desc = cl_getmsg('lib-00212',g_rlang) 
      #180724-00009#2 add(s)
      #180724-00009#2 add(e)
      WHEN "7" #明細清單(多行單身)
         LET l_gzgf006_desc = cl_getmsg('lib-00639',g_rlang)
      #180724-00009#2 add(e)         
    END CASE
      
   ###樣板名稱
   SELECT gzgdl002 INTO l_gzgf000_desc FROM gzgdl_t
    WHERE gzgdl000 = p_gzgf000 AND gzgdl001 = g_lang   
   
   ###用戶說明
   SELECT ooag011 INTO l_gzgf004_desc FROM ooag_t 
    WHERE ooagstus = "Y" AND ooagent = g_enterprise
      AND ooag001 = p_gzgf004

   IF cl_null(l_gzgf004_desc) THEN
      SELECT gzze003 INTO l_gzgf004_desc FROM gzze_t 
       WHERE gzzestus = "Y" AND gzze001 = 'azz-00665' AND gzze002 = g_lang
   END IF
   IF cl_null(l_gzgf004_desc) THEN
      LET l_gzgf004_desc = p_gzgf004
   END IF   
   
   ###角色說明
   SELECT gzyal003 INTO l_gzgf005_desc FROM gzyal_t 
    WHERE gzyalent = g_enterprise AND gzyal001 = p_gzgf005
      AND gzyal002 = g_lang  

   IF cl_null(l_gzgf005_desc) THEN
      SELECT gzze003 INTO l_gzgf005_desc FROM gzze_t 
       WHERE gzzestus = "Y" AND gzze001 = 'azz-00666' AND gzze002 = g_lang
   END IF
   IF cl_null(l_gzgf005_desc) THEN
      LET l_gzgf005_desc = p_gzgf005
   END IF   
   LET r_gzgh011 = l_gzgf006_desc,";",l_gzgf000_desc,";",l_gzgf004_desc,";",l_gzgf005_desc,";",p_gzgf002
   #LET r_gzgh011 = l_gzgf006_desc

   RETURN r_gzgh011
END FUNCTION

################################################################################
# Descriptions...: 組合欄位說明gzgh041的值
# Memo...........:
# Usage..........: CALL cl_xg_get_gzgh041 (p_gzgh002)
#                  RETURNING 回传参数
# Input parameter: p_gzgh002      報表樣板ID  
#                : 
# Return code....: l_gzgh041 
#                : 
# Date & Author..: 140703 By 02716
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_get_gzgh041(p_gzgh002)
   DEFINE p_gzgh002	LIKE gzgh_t.gzgh002
   DEFINE l_gzgh041	STRING       
   DEFINE l_gzgg001	LIKE gzgg_t.gzgg001   #欄位代碼
   DEFINE l_gzge003	LIKE gzge_t.gzge003   #欄位說明       
   DEFINE l_sql 	   STRING
   DEFINE l_gzgg004  LIKE gzgg_t.gzgg004 #150413 add
   DEFINE l_cnt      LIKE gzgg_t.gzgg004 #150413 add
   DEFINE l_i        LIKE type_t.num5    #150413 add
   DEFINE l_id       LIKE gzgg_t.gzgg004 #150413 add
   DEFINE l_db       STRING
 
   #LOCATE l_gzgh041 IN FILE
   LET l_gzgh041 = "" 
   
#150413 add (s)
   LET l_sql = "SELECT gzgg001 ", 
               "  FROM gzgg_t ",
               " WHERE gzgg000 = '",p_gzgh002,"'",
               "   AND gzgg002 = '",g_rlang,"'", 
               #"   AND ( gzgg019 = '2' OR gzgg019 IS NULL) ",   #141020 單頭 add   #141021 mark
               #" ORDER BY gzgg004 "   #141021 單頭mark  
               " ORDER BY gzgg019,gzgg027,gzgg004 "   #141021 單頭add  
  
   PREPARE gzgh041_prepare FROM l_sql
   DECLARE gzgh041_cursor CURSOR FOR gzgh041_prepare
      
   LET l_db = g_rep_db CLIPPED
   LET l_db = cl_replace_str(l_db,".","")
   IF g_xg_fieldname.getLength() > 0 THEN
      #PGS(D)-05292
      LET l_sql = " SELECT column_id FROM all_tab_columns",
                    "  WHERE lower(owner)='",l_db CLIPPED,"' ",
                    "    AND lower(table_name)='",g_rep_tmpname[1] CLIPPED,"'",
                  "    AND lower(column_name) = ? "
      #PGS(D)-05292
      PREPARE field_name_prepare FROM l_sql
      DECLARE field_name_cursor CURSOR FOR field_name_prepare
      
      #171205-00041#1 add(s)
      #子報表欄位清單
      #PGS(D)-05292
      LET l_sql = " SELECT column_id FROM all_tab_columns",
                    "  WHERE lower(owner)='",l_db CLIPPED,"' ",
                    "    AND lower(table_name)='",g_rep_tmpname[2] CLIPPED,"'",
                  "    AND lower(column_name) = ? "
      #PGS(D)-05292
      PREPARE sub_field_name_prepare FROM l_sql
      DECLARE sub_field_name_cursor CURSOR FOR sub_field_name_prepare      
      #171205-00041#1 add(e)
      
      FOREACH gzgh041_cursor INTO l_gzgg001
         LET l_id = ""      #171205-00041#1 add
         EXECUTE field_name_cursor INTO l_id USING l_gzgg001
         LET l_gzge003 = ""
         IF NOT cl_null(l_id) AND l_id > 0 THEN
            LET l_gzge003 = g_xg_fieldname[l_id]
         #171205-00041#1 add(s)
         ELSE
            #有子報表
            IF NOT cl_null(g_rep_tmpname[2]) THEN
               EXECUTE sub_field_name_cursor INTO l_id USING l_gzgg001
               LET l_gzge003 = g_xg_fieldname1[l_id]
            END IF
         #171205-00041#1 add(e)
         END IF
         
         IF cl_null(l_gzge003) THEN   
            SELECT gzge003 INTO l_gzge003
              FROM gzge_t
             WHERE gzge000 = p_gzgh002
               AND gzge001 = l_gzgg001 
               AND gzge002 = g_rlang
               AND gzgestus = 'Y'
         END IF            
   
         IF cl_null(l_gzge003) THEN     
            SELECT dzebl003 INTO l_gzge003
              FROM dzebl_t
             WHERE dzebl001 = l_gzgg001 
               AND dzebl002 = g_rlang
         END IF   
   
         IF cl_null(l_gzge003) OR l_gzge003 = "" THEN
            LET l_gzge003 = l_gzgg001
         END IF          

         LET l_gzgh041 = l_gzgh041 CLIPPED, l_gzgg001,":", l_gzge003,"|"
      END FOREACH
   ELSE                
#150413 add (e)         
#      LET l_sql = "SELECT gzgg001 ", 
#                  "  FROM gzgg_t ",
#                  " WHERE gzgg000 = '",p_gzgh002,"'",
#                  "   AND gzgg002 = '",g_rlang,"'", 
#                  #"   AND ( gzgg019 = '2' OR gzgg019 IS NULL) ",   #141020 單頭 add   #141021 mark
#                  #" ORDER BY gzgg004 "   #141021 單頭mark  
#                  " ORDER BY gzgg019,gzgg027,gzgg004 "   #141021 單頭add  
#   
#      PREPARE gzgh041_prepare FROM l_sql
#      DECLARE gzgh041_cursor CURSOR FOR gzgh041_prepare
   
      FOREACH gzgh041_cursor INTO l_gzgg001
         LET l_gzgh041 = l_gzgh041 CLIPPED, l_gzgg001,":"
         LET l_gzge003 = ""
         SELECT gzge003 INTO l_gzge003
           FROM gzge_t
          WHERE gzge000 = p_gzgh002
            AND gzge001 = l_gzgg001 
            AND gzge002 = g_rlang
            AND gzgestus = 'Y' 
   
         IF cl_null(l_gzge003) THEN     
            SELECT dzebl003 INTO l_gzge003
              FROM dzebl_t
             WHERE dzebl001 = l_gzgg001 
               AND dzebl002 = g_rlang
         END IF   
   
         IF cl_null(l_gzge003) OR l_gzge003 = "" THEN
            LET l_gzge003 = l_gzgg001
         END IF        
   
         LET l_gzgh041 = l_gzgh041 CLIPPED, l_gzge003,"|" 
      END FOREACH
   END IF  #150413 add
   
   FREE gzgh041_cursor
   FREE field_name_cursor
   LET l_gzgh041 = l_gzgh041.substring(1,LENGTH(l_gzgh041)-1) 
   #DISPLAY "l_gzgh041:",l_gzgh041
   RETURN l_gzgh041 
   #FREE l_gzgh041 
END FUNCTION

################################################################################
# Descriptions...: 取得表頭表尾資訊
# Memo...........:
# Usage..........: CALL cl_xg_get_gzgh046()
#                  RETURNING r_gzgh046
# Input parameter: NULL
# Return code....: r_gzgh046      表頭表尾資訊
# Date & Author..: 2014/11/6 By 04010
# Modify.........: 150904-00005#4 加入表尾資訊
################################################################################
PUBLIC FUNCTION cl_xg_get_gzgh046()
   DEFINE r_gzgh046     STRING
   DEFINE l_gzgm003     LIKE gzgm_t.gzgm003    #表頭代號
   DEFINE l_gzgk006     LIKE gzgk_t.gzgk006    #顯示內容 
   DEFINE l_ooek004     LIKE ooek_t.ooek004    #列印國家地區
   DEFINE l_ooek005     LIKE ooek_t.ooek005    #列印州/省/直轄市   
   DEFINE l_title       STRING
   DEFINE l_sql         STRING
   DEFINE l_cnt         INTEGER                #紀錄COUNT 結果  #180524-00013#1 add

   LET r_gzgh046 = ""
   
   #地址是否列印國家地區(ooek004),州/省/直轄市(ooek005)
   SELECT ooek004,ooek005 INTO l_ooek004,l_ooek005 FROM ooek_t
    WHERE ooekstus = "Y" AND ooekent = g_enterprise AND ooek001 = g_site AND ooek002 = g_rlang
   
   IF cl_null(l_ooek004) THEN LET l_ooek004 = "Y" END IF
   
   IF cl_null(l_ooek005) THEN LET l_ooek005 = "Y" END IF
   
   #取得表頭代號,預設為"YY"
   SELECT MAX(gzgm003) INTO l_gzgm003 FROM gzgm_t
    WHERE gzgmstus = "Y" AND gzgm001 = g_rep_code AND gzgm002 = g_dlang

#   IF cl_null(l_gzgm003) THEN LET l_gzgm003 = "YY" END IF   #180524-00013#1 mark

   #180524-00013#1 add(s)
   #組好SQL指令
   LET l_sql = "SELECT COUNT(1) FROM gzgj_t WHERE gzgj001 = 'T100_YY'"
   #準備CURSOR
   PREPARE cl_gr_init_curs FROM l_sql
   #執行CURSOR
   EXECUTE cl_gr_init_curs INTO l_cnt

   #判斷如果l_gzgm003為空值就往下執行
   IF cl_null(l_gzgm003) THEN
      #判斷COUNT不等於0就帶入T100_YY，若等於0就帶入YY
      IF l_cnt <> 0 THEN
         LET l_gzgm003 = "T100_YY"
      ELSE
         LET l_gzgm003 = "YY"
      END IF
   END IF
   #180524-00013#1 add(e)
   
   CALL cl_gr_get_addr(l_gzgm003,l_ooek004,l_ooek005)   #地址   
   
   #取得表頭表尾設定,如："11:鼎捷集團台中分公司|16:銷售金額統計表"
   LET l_sql = "SELECT gzgk006 FROM gzgk_t ",
               " WHERE gzgk001 = '",l_gzgm003,"' AND gzgk002 = '",g_dlang,"' ",
#               "  AND gzgk014 = '1' ",   #150611-00032#4 add   #150904-00005#4 mark
#               " ORDER BY gzgk003,gzgk004 "   #150904-00005#4 mark
               " ORDER BY gzgk006 "   #150904-00005#4 add
   
   DECLARE cl_xg_gzgk006_cur CURSOR FROM l_sql
 
   FOREACH cl_xg_gzgk006_cur INTO l_gzgk006
      IF NOT cl_null(l_gzgk006) THEN
         #160217-00006#1 add(s)
         IF l_gzgk006 = '16' AND NOT cl_null(g_xg_rep_title) THEN
            LET l_title = g_xg_rep_title CLIPPED
         ELSE   
            LET l_title = cl_gr_title(l_gzgk006)
         END IF
         #160217-00006#1 add(e)
#         LET l_title = cl_gr_title(l_gzgk006)   #160217-00006#1 mark
         
         #150904-00005#4 add(s)
         IF l_gzgk006 = "23" THEN LET l_title = cl_getmsg("azz-00903",g_dlang) CLIPPED END IF
         IF l_gzgk006="20" THEN   #製表日期與時間
#            LET l_title = cl_getmsg("lib-00395",g_dlang) CLIPPED,": ",TODAY USING g_date_fmt, " ",cl_get_time()    #161117-00037#1 mark
            LET l_title = cl_getmsg("lib-00395",g_dlang) CLIPPED,": ",g_ldate USING g_date_fmt, " ",cl_get_time()   #161117-00037#1 add
         END IF
         IF l_gzgk006="21" THEN   #製表日期 
#            LET l_title = cl_getmsg("lib-00146",g_dlang) CLIPPED,TODAY USING g_date_fmt    #161117-00037#1 mark
            LET l_title = cl_getmsg("lib-00146",g_dlang) CLIPPED,g_ldate USING g_date_fmt   #161117-00037#1 add
         END IF            
         #150904-00005#4 add(e)         
         
         LET r_gzgh046 = r_gzgh046,"|",l_gzgk006 CLIPPED,":",l_title CLIPPED      
      END IF   
   END FOREACH

   LET r_gzgh046 = r_gzgh046.subString(2,r_gzgh046.getLength())
   
   #170828-00025#1 add(s)
   #如果抓不到表頭表尾資訊則使用固化YY的內容，避免多語言資料未建立導致報表無法列印
   #範例:11:鼎新電腦|16:料件多營運據點庫存查詢|17:(ainr100)|20:製表日期與時間: 2017/09/25 15:47:30|22:製表者: tiptop|23:頁次
   IF cl_null(r_gzgh046) THEN
      LET l_gzgm003 = "YY"
      LET r_gzgh046 = "11:",cl_gr_title("11"),"|16:"
      IF NOT cl_null(g_xg_rep_title) THEN
         LET r_gzgh046 = r_gzgh046,g_xg_rep_title CLIPPED
      ELSE   
         LET r_gzgh046 = r_gzgh046,cl_gr_title("16")
      END IF
      LET r_gzgh046 = r_gzgh046,"|17:",cl_gr_title("17"),"|20:",cl_getmsg("lib-00395",g_dlang) CLIPPED,": ",g_ldate USING g_date_fmt, " ",cl_get_time(),
                      "|22:",cl_gr_title("22"),"|23:",cl_getmsg("azz-00903",g_dlang)
   END IF
   #170828-00025#1 add(e)   
   
   RETURN r_gzgh046 ,l_gzgm003
END FUNCTION

################################################################################
# Descriptions...: 取得不可視欄位代碼 column_visible
# Memo...........:
# Usage..........: CALL cl_xg_get_gzgh047()
#                  RETURNING r_gzgh047
# Input parameter: NULL
# Return code....: r_gzgh047      不可視欄位代碼
# Date & Author..: 2014/11/6 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_get_gzgh047()
   DEFINE r_gzgh047     STRING
   DEFINE l_gzza002     LIKE gzza_t.gzza002   #程式類別(T/R)
   DEFINE l_gzyd003     LIKE gzyd_t.gzyd003   #欄位編號   
   DEFINE l_prog        STRING
   DEFINE l_sql         STRING
#   DEFINE l_roles       STRING               #150709-00014#1 mark
   DEFINE lst_token     base.StringTokenizer  #150709-00014#1 add
   DEFINE ls_token      STRING                #150709-00014#1 add
   DEFINE l_str         STRING                #170621-00048#2 add

   LET l_prog = ""
   LET r_gzgh047 = ""
#   LET l_roles = cl_replace_str(g_roles,",","','")   #150709-00014#1 mark
   
   #取得程式類別
   SELECT gzza002 INTO l_gzza002 FROM gzza_t
    WHERE gzza001 = g_code   #170621-00048#2 add   #azzi900的資料要用g_code取得(ex:axrq380)
#    WHERE gzza001 = g_prog   #170621-00048#2 mark

#   IF l_gzza002 = "R" THEN  #R類直接撈azzi850的資料   #170621-00048#2 mark
   IF l_gzza002 = "R" OR l_gzza002 = "Q" THEN  #R類直接撈azzi850的資料 #170621-00048#2 add
      LET l_prog = g_prog                      #程式代號要以當下執行代號為主(ex:aapq380)
   ELSE
      #T類利用報表元件代號去掉後四碼(_x01)撈azzi850的資料
      LET l_prog = g_rep_code
      LET l_prog = l_prog.subString(1,l_prog.getLength()-4)      
   END IF
   
   #取得不可視欄位
   #150709-00014#1 mark(s)
#   LET l_sql = " SELECT gzyd003 FROM gzyd_t ",
#               " WHERE gzydent = '",g_enterprise,"' ", "AND gzyd001 IN ('",l_roles CLIPPED,"')", #"AND gzyd001 IN '",g_roles,"' ",
#               " AND gzyd002 = '",l_prog,"' ",
#               " ORDER BY gzyd003 " 
   #150709-00014#1 mark(e)
   
   #150709-00014#1 add(s)
   #組合SQL(撈取設定隱藏的欄位)
   LET lst_token = base.StringTokenizer.create(g_roles, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()

      IF NOT cl_null(l_sql) THEN
         LET l_sql = l_sql, " INTERSECT "
      END IF

      LET l_sql = l_sql,
                   " SELECT gzyd003 FROM gzyd_t ",
                   "  WHERE gzydent = '",g_enterprise,"'", #企業代碼
                   "    AND gzyd001 = '",ls_token,"'",     #職能腳色編號
                   "    AND gzyd002 = '",l_prog,"'",       #程式名稱
                   "    AND gzyd004 = '1' "                #欄位隱藏
   END WHILE   
   #150709-00014#1 add(e)

   DECLARE cl_xg_gzyd003_cur CURSOR FROM l_sql
 
   FOREACH cl_xg_gzyd003_cur INTO l_gzyd003
      IF NOT cl_null(l_gzyd003) THEN
         #170621-00048#2 add(s)
         LET l_str = l_gzyd003
         IF l_str.getIndexOf(".",1) = 0 OR l_str.getIndexOf(g_rep_code,1) > 0 THEN
            LET l_gzyd003 = cl_str_replace(l_str,g_rep_code||".","")
         #170621-00048#2 add(e)
            LET r_gzgh047 = r_gzgh047,"|",l_gzyd003 CLIPPED
         END IF      #170621-00048#2 add
      END IF
   END FOREACH

   IF cl_null(g_xgrid.visible_column) THEN
      LET r_gzgh047 = r_gzgh047.subString(2,r_gzgh047.getLength())
   ELSE
      LET r_gzgh047 = g_xgrid.visible_column CLIPPED,r_gzgh047
   END IF

   RETURN r_gzgh047 
END FUNCTION

################################################################################
# Descriptions...: 查詢報表直接送印
# Memo...........:
# Usage..........: CALL cl_xg_direct_print()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2014/12/9 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_direct_print()
   DEFINE i,j,l_i,l_j,l_ac         LIKE type_t.num5
   DEFINE l_index              DYNAMIC ARRAY WITH DIMENSION 2 OF STRING
   DEFINE l_sql                STRING
   DEFINE l_str                STRING
   DEFINE l_tok_param          base.StringTokenizer
   DEFINE l_roles              STRING
   DEFINE l_roles_arr          DYNAMIC ARRAY OF STRING
   DEFINE l_dp_tmp             DYNAMIC ARRAY OF type_g_gzgn  #記錄直接送印資料(暫存)
   DEFINE l_dp                 DYNAMIC ARRAY OF type_g_gzgn  #記錄直接送印資料
   DEFINE l_dp_i               LIKE type_t.num5              #使用第幾筆直接送印的資料
   DEFINE l_rootnode	          om.DomNode
   DEFINE l_nodes	             om.NodeList
   DEFINE l_dialog             ui.Dialog
   DEFINE l_gzgn004            STRING
   DEFINE l_client_ip          LIKE type_t.chr20             #Client端IP
   DEFINE l_sort               LIKE type_t.num5              #Client端IP的優先序，若為6代表不符合直接送印設定的IP
   DEFINE l_flag               LIKE type_t.num5
   
   WHENEVER ERROR CONTINUE
   
   LET l_roles = cl_replace_str(g_roles,",","','")
   
   LET l_sql = "SELECT gzgn001,gzgn002,gzgn003,gzgn004,gzgn005,gzgn006,gzgn007,gzgn008,gzgn009,gzgn010",
               " FROM gzgn_t WHERE gzgn003 = '",g_prog CLIPPED,"'",
               " AND ((gzgn001 = 'default' AND gzgn002 = 'default')",
               " OR gzgn001 = '",g_user CLIPPED,"' OR gzgn002 IN ('",l_roles CLIPPED,"'))",
               " AND gzgnstus = 'Y' ORDER BY gzgn001,gzgn002,gzgn004,gzgn005"
   
   DECLARE cl_xg_dp_curs CURSOR FROM l_sql
   
   CALL l_dp_tmp.clear()
   LET l_i = 1
   FOREACH cl_xg_dp_curs INTO l_dp_tmp[l_i].*
      IF cl_null(l_dp_tmp[l_i].gzgn007) OR l_dp_tmp[l_i].gzgn007 < 1 THEN
         LET l_dp_tmp[l_i].gzgn007 = 1
      END iF
      
      LET l_i = l_i + 1
   END FOREACH
   CLOSE cl_xg_dp_curs
   FREE cl_xg_dp_curs
   
   IF l_dp_tmp.getLength() >= 1 THEN
      CALL l_dp_tmp.deleteElement(l_i)
   END IF
    
   #將多個職能角色編號存入陣列中
   LET l_tok_param = base.StringTokenizer.createExt(g_roles,",","",TRUE)
   CALL l_roles_arr.clear()
   LET i = 0
   WHILE l_tok_param.hasMoreTokens()
      LET i = i + 1
      LET l_roles_arr[i] = l_tok_param.nextToken()
   END WHILE
   
   #決定直接送印的優先序：g_user > g_roles > default
   FOR l_i = 1 TO l_dp_tmp.getLength()
      LET l_str = l_dp_tmp[l_i].gzgn002   #職能角色編號
      LET l_client_ip = cl_getenv_client_ip()
      CASE
         #將用戶設定的直接送印設定資料集中
         WHEN l_dp_tmp[l_i].gzgn001 = g_user
            #Client端IP須符合直接送印設定的IP，並決定優先序
            #IP優先序預設為6，代表直接送印設定的IP(p_ip)與Client端IP(p_client_ip)比對不符合
            LET l_sort = cl_gr_direct_print_chkip(l_dp_tmp[l_i].gzgn010,l_client_ip)
            IF l_sort != 6 THEN
               #若有多筆直接送印設定，以「,」分隔
               IF l_index[1,l_sort] IS NULL THEN
                  LET l_index[1,l_sort] = l_i
               ELSE
                  LET l_index[1,l_sort] = l_index[1,l_sort],",",l_i
               END IF
            END IF
         #將角色設定的直接送印設定資料集中
         WHEN g_roles.getindexof(l_str, 1) > 0
            FOR i = 1 TO l_roles_arr.getLength()
               IF l_str.equals(l_roles_arr[i]) THEN
                  LET l_sort = cl_gr_direct_print_chkip(l_dp_tmp[l_i].gzgn010,l_client_ip)
                  IF l_sort != 6 THEN
                     IF l_index[2,l_sort] IS NULL THEN
                        LET l_index[2,l_sort] = l_i
                     ELSE
                        LET l_index[2,l_sort] = l_index[2,l_sort],",",l_i
                     END IF
                  END IF
               END IF
            END FOR
         #將default/default設定的直接送印設定資料集中
         WHEN l_dp_tmp[l_i].gzgn001 = 'default' AND l_dp_tmp[l_i].gzgn002 = 'default'
            LET l_sort = cl_gr_direct_print_chkip(l_dp_tmp[l_i].gzgn010,l_client_ip)
            IF l_sort != 6 THEN
               IF l_index[3,l_sort] IS NULL THEN
                  LET l_index[3,l_sort] = l_i
               ELSE
                  LET l_index[3,l_sort] = l_index[3,l_sort],",",l_i
               END IF
            END IF
      END CASE

      SELECT gzgf000,gzgf001,gzgf002,gzgf003,gzgf004,gzgf005,gzgf008,gzgf009,gzgf010
      INTO l_dp_tmp[l_i].gzgf000,l_dp_tmp[l_i].gzgf001,l_dp_tmp[l_i].gzgf002,l_dp_tmp[l_i].gzgf003,
           l_dp_tmp[l_i].gzgf004,l_dp_tmp[l_i].gzgf005,l_dp_tmp[l_i].gzgf008,l_dp_tmp[l_i].gzgf009,
           l_dp_tmp[l_i].gzgf010
      FROM gzgf_t
      WHERE gzgf002 = l_dp_tmp[l_i].gzgn004
      
      SELECT gzgdl002 INTO l_dp_tmp[l_i].gzgdl002 
      FROM gzgdl_t 
      WHERE gzgdl000 = l_dp_tmp[l_i].gzgf000 AND gzgdl001 = g_dlang
   END FOR

   #若有多筆直接送印設定，開窗讓使用者選擇樣板
   #再將直接送印的資料填入全域變數(g_dp)中
   FOR l_i = 1 TO l_index.getLength()
      FOR l_j = 1 TO l_index[l_i].getLength()
         IF l_index[l_i,l_j] IS NOT NULL THEN
            #將直接送印設定存入區域變數(l_dp)中
            IF l_index[l_i,l_j].getindexof(",", 1) > 0 THEN
               LET l_tok_param = base.StringTokenizer.createExt(l_index[l_i,l_j],",","",TRUE)
               LET i = 1
               LET l_dp_i = 1
               WHILE l_tok_param.hasMoreTokens()
                  LET l_dp_i = l_tok_param.nextToken()
                  LET l_dp[i].* = l_dp_tmp[l_dp_i].*
                  
                  #開窗會使用到的全域變數(g_gzgd)
                  LET g_gzgf[i].pick = 'N'
                  LET g_gzgf[i].gzgdl002 = l_dp[i].gzgdl002
                  LET g_gzgf[i].gzgf003 = l_dp[i].gzgf003
                  LET g_gzgf[i].gzgf002 = l_dp[i].gzgf004
                  LET g_gzgf[i].gzgf002 = l_dp[i].gzgf002
                  LET g_gzgf[i].gzgf000 = l_dp[i].gzgf000
                  LET g_gzgf[i].gzgf005 = l_dp[i].gzgf005
                  LET g_gzgf[i].gzgf004 = l_dp[i].gzgf004
                  LET g_gzgf[i].gzgf007 = l_dp[i].gzgf002   #170622-00040#1 add
                  LET g_gzgf[i].gzgf009 = l_dp[i].gzgf009
                  LET g_gzgf[i].gzgf010 = l_dp[i].gzgf010
                  LET g_gzgf[i].sort_num = i
                  
                  LET i = i + 1
               END WHILE
            ELSE
               LET l_dp_i = l_index[l_i,l_j]
               LET l_dp[1].* = l_dp_tmp[l_dp_i].*
               #LET g_dp.* = l_dp_tmp[l_dp_i].*
            END IF
            EXIT FOR
         END IF
      END FOR   
      #170622-00040#1 add(s)
      IF l_index[l_i,l_j] IS NOT NULL THEN
         EXIT FOR
      END IF
      #170622-00040#1 add(e)
   END FOR
 
   #若區域變數(l_dp)多於1筆，開窗讓使用者選擇樣板
   IF l_dp.getLength() > 1 THEN
      #170622-00040#1 mark(s)
      #LET l_rootnode = ui.Interface.getRootNode()
      #LET l_nodes = l_rootnode.selectByPath("//Window[@name=\"cl_gr_w\"]")
      #IF l_nodes.getLength() <= 0 THEN
      #170622-00040#1 mark(e)
         OPEN WINDOW cl_gr_w WITH FORM cl_ap_formpath("lib","cl_gr_s01")         
            ATTRIBUTES(STYLE="layoutwin") 
      #END IF  #170622-00040#1 mark
      
      CALL cl_ui_init()
      
      CALL cl_set_comp_visible("layout_edit",FALSE)
      CALL cl_set_comp_visible("sampledata",FALSE)
      CALL cl_set_comp_visible("outputformat",FALSE)
      CALL cl_set_comp_visible("set_default",FALSE)
      
      DIALOG ATTRIBUTE(UNBUFFERED)
         DISPLAY ARRAY g_gzgf TO s_gzgd.* ATTRIBUTE(COUNT = g_rec_b)
            BEFORE DISPLAY
               LET l_i = 1
               LET g_gzgf[1].pick = 'Y'                 #初始狀態第一行必勾選
               IF g_gzgf.getLength() != 1 THEN          #若不只一筆樣板檔
                  FOR l_i = 2 TO g_gzgf.getLength()     #其餘筆數均預設不勾選
                     LET g_gzgf[l_i].pick = 'N'
                  END FOR
                  CALL g_gzgf.deleteElement(l_i)        #刪除系統自動產生的最後一筆空行
                  LET l_ac = 1
               END IF
            
            BEFORE ROW
               IF g_gzgf.getLength() != 1 AND l_ac != 0 THEN       #若不只一筆樣板檔
                  IF l_ac != l_dialog.getCurrentRow("s_gzgd") THEN #選擇新筆資料時
                     LET g_gzgf[l_ac].pick = 'N'                   #將上筆取消勾選
                     LET l_ac = l_dialog.getCurrentRow("s_gzgd")   #取得新筆行數
                     LET g_gzgf[l_ac].pick = 'Y'                   #新筆設置為勾選
                  END IF
               END IF
            
            AFTER ROW
               LET l_ac = l_dialog.getCurrentRow("s_gzgd")
         END DISPLAY
         
         BEFORE DIALOG
            LET l_dialog = ui.Dialog.getCurrent()
         
         ON IDLE g_idle_seconds
            CALL cl_on_idle()
         
         ON ACTION btn_accept
            ACCEPT DIALOG
         
         ON ACTION btn_cancel
            LET INT_FLAG = TRUE
            EXIT DIALOG
         
         ON ACTION CLOSE
            LET INT_FLAG = TRUE
            EXIT DIALOG
      END DIALOG
      
      #將選取到的樣板存入全域變數(g_dp)
      FOR i = 1 TO g_gzgf.getLength()
         IF g_gzgf[i].pick = 'Y' THEN
            LET l_gzgn004 = g_gzgf[i].gzgf002
         END IF
      END FOR
      
      #FOR i = 1 TO l_dp_tmp.getLength()  #170622-00040#1 mark
      FOR i = 1 TO l_dp.getLength()    #170622-00040#1 add
         IF l_gzgn004.equals(l_dp[i].gzgn004) THEN
            LET g_dp.* = l_dp[i].*
         END IF
      END FOR
      #畫面關閉
      CLOSE WINDOW cl_gr_w    #170622-00040#1 add
   ELSE
      LET g_dp.* = l_dp[1].*
   END IF

   #使用者/角色是否有設定直接送印
   LET g_dp_flag = 'N'
   IF l_dp_i >= 1 THEN
      LET g_dp_flag = 'Y'
   END IF
                                                                                                                                                                                                                                                                                                                                                 
END FUNCTION

################################################################################
# Descriptions...: 檢查欄位個數
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_n            传入参数变量说明1
#                : p_gzgf000      传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2015/4/14 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_chk_tmp_table_field_1(p_n,p_gzgf000)
   DEFINE p_n          LIKE type_t.num10
   DEFINE p_gzgf000    LIKE gzgf_t.gzgf000
   DEFINE l_sql        STRING
   DEFINE l_i          LIKE type_t.num5 
   DEFINE l_gzgg017    LIKE gzgg_t.gzgg017

   TYPE type_field_chk DYNAMIC ARRAY OF RECORD
          gzgg001      LIKE gzgg_t.gzgg001,   #報表欄位代碼
          gzgg003      LIKE gzgg_t.gzgg003,   #欄位型態 C/N/D
          flag         LIKE type_t.chr1       #是否已比對過
   
         END RECORD 
   
   DEFINE l_gzgg       type_field_chk
   
#   DEFINE l_gzgg   DYNAMIC ARRAY OF RECORD
#          gzgg001      LIKE gzgg_t.gzgg001,   #報表欄位代碼
#          gzgg003      LIKE gzgg_t.gzgg003,   #欄位型態 C/N/D
#          flag         LIKE type_t.chr1       #是否已比對過
#          END RECORD
          
          
   #DEFINE l_field_list DYNAMIC ARRAY OF LIKE gzgg_t.gzgg001 #temp table欄位清單          
   DEFINE r_msg        STRING
   
#   IF p_n = 1 THEN   #主報表
#      LET l_sql = "SELECT gzgg001 FROM gzgg_t WHERE gzgg000 = '",p_gzgf000 CLIPPED,"' AND gzgg002 = '",g_rlang CLIPPED,"' AND (gzgg017 IS NULL OR gzgg017 = 'Y')"
#      LET r_msg = cl_getmsg("lib-00241",g_lang) CLIPPED
#      LET l_i = 0
#   ELSE
#      LET l_gzgg017 = p_n
#      LET l_sql = "SELECT gzgg001 FROM gzgg_t WHERE gzgg000 = '",p_gzgf000 CLIPPED,"' AND gzgg002 = '",g_rlang CLIPPED,"' AND gzgg017 = '",l_gzgg017 CLIPPED,"'"           
#      LET r_msg = cl_getmsg("lib-00242",g_lang) CLIPPED   
#   END IF
   
   
   
   
END FUNCTION

################################################################################
# Descriptions...: 檢查暫存表欄位不一致
# Memo...........:
# Usage..........: CALL cl_xg_chk_tmp_table_field(p_flag,p_n,p_gzgf000)
#                  RETURNING r_msg
# Input parameter: p_flag         欄位個數大於或小於
#                : p_n            暫存表序號
#                : p_gzgf000      樣板ID
# Return code....: r_msg          欄位清單
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2015/1/9 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_chk_tmp_table_field(p_flag,p_n,p_gzgf000)
   DEFINE p_flag       LIKE type_t.chr1
   DEFINE p_n          LIKE type_t.num10
   DEFINE p_gzgf000    LIKE gzgf_t.gzgf000
   DEFINE l_sql        STRING
   DEFINE l_gzgg001    LIKE gzgg_t.gzgg001
   DEFINE l_gzgg017    LIKE gzgg_t.gzgg017
   DEFINE l_gzgh003    LIKE gzgh_t.gzgh003
   DEFINE l_field_cnt  LIKE type_t.num5 
   DEFINE r_msg        STRING
   DEFINE l_report     STRING
   
   LET l_report = g_report
   
   LET r_msg = NULL
   #報表元件中create temp table的個數 < azzi300中設定的欄位個數
   IF p_flag = "l" THEN
      IF p_n = 1 THEN   #主報表
         LET l_sql = "SELECT gzgg001 FROM gzgg_t WHERE gzgg000 = '",p_gzgf000 CLIPPED,"' AND gzgg002 = '",g_rlang CLIPPED,"' AND (gzgg017 IS NULL OR gzgg017 = 'Y' OR gzgg017 = 'N')"
         LET r_msg = cl_getmsg("lib-00241",g_lang) CLIPPED
      ELSE
         LET l_gzgg017 = p_n
         LET l_sql = "SELECT gzgg001 FROM gzgg_t WHERE gzgg000 = '",p_gzgf000 CLIPPED,"' AND gzgg002 = '",g_rlang CLIPPED,"' AND gzgg017 = '",l_gzgg017 CLIPPED,"'"           
         LET r_msg = cl_getmsg("lib-00242",g_lang) CLIPPED
      END IF 
      
      DECLARE cl_xg_chk_field_cs1 CURSOR FROM l_sql
      LET l_gzgh003 = g_rep_tmpname[p_n]
      FOREACH cl_xg_chk_field_cs1 INTO l_gzgg001
         SELECT COUNT(COLUMN_NAME) INTO l_field_cnt FROM  all_tab_columns 
          WHERE lower(table_name) = l_gzgh003 AND lower(owner) = g_report AND lower(COLUMN_NAME) = l_gzgg001
          
         IF l_field_cnt < 1 THEN
            LET r_msg = r_msg,l_gzgg001,","                            
         END IF         
      END FOREACH      
   #報表元件中create temp table的個數 > azzi300中設定的欄位個數
   ELSE
      #PGS(D)-05292
      #240130-00011#2 mark(s)
      #LET l_sql = "SELECT COLUMN_NAME FROM  all_tab_columns WHERE lower(table_name) = '",
      #            g_rep_tmpname[p_n] CLIPPED ,"' AND lower(owner) = '",g_report CLIPPED,"'"
      #240130-00011#2 mark(e)
      #240130-00011#2 add(s)
      LET l_sql = "SELECT COLUMN_NAME ",
                  "  FROM all_tab_columns ",
                  " WHERE     table_name = '", g_rep_tmpname[p_n].toUpperCase() CLIPPED,"' ",
                  "       AND owner = '", l_report.toUpperCase() CLIPPED, "' "
      #240130-00011#2 add(e)
      #PGS(D)-05292
      
      DECLARE cl_xg_chk_field_cs2 CURSOR FROM l_sql
      
      FOREACH cl_xg_chk_field_cs2 INTO l_gzgg001
         LET l_gzgg001 = DOWNSHIFT(l_gzgg001)      
         IF p_n = 1 THEN   #主報表
            SELECT COUNT(gzgg001) INTO l_field_cnt FROM gzgg_t 
             WHERE gzgg000 = p_gzgf000 AND gzgg001 = l_gzgg001 AND gzgg002 = g_rlang 
#             WHERE gzgg000 = p_gzgf000 AND gzgg001 = lower(l_gzgg001) AND gzgg002 = g_rlang 
#               AND (gzgg017 IS NULL OR gzgg017 = 'Y')   #150930-00022#2 mark
               AND (gzgg017 IS NULL OR gzgg017 = 'Y' OR gzgg017 = 'N')   #150930-00022#2 add
            IF l_field_cnt < 1 THEN
               IF cl_null(r_msg) THEN            
                  LET r_msg = cl_getmsg("lib-00241",g_lang) CLIPPED,l_gzgg001
               ELSE
                  LET r_msg = r_msg,",",l_gzgg001
               END iF   
            END IF            
         ELSE
            LET l_gzgg017 = p_n
            LET l_gzgg017 = l_gzgg017 CLIPPED
            SELECT COUNT(gzgg001) INTO l_field_cnt FROM gzgg_t 
             WHERE gzgg000 = p_gzgf000 AND gzgg001 = l_gzgg001 AND gzgg002 = g_rlang AND gzgg017 = l_gzgg017
#             WHERE gzgg000 = p_gzgf000 AND gzgg001 = lower(l_gzgg001) AND gzgg002 = g_rlang AND gzgg017 = l_gzgg017
            IF l_field_cnt < 1 THEN  
               IF cl_null(r_msg) THEN 
                  LET r_msg = cl_getmsg("lib-00242",g_lang) CLIPPED,l_gzgg001,","
               ELSE
                  LET r_msg = r_msg,l_gzgg001,","
               END IF                 
            END IF                        
         END IF         
      END FOREACH    
   END IF   
   
   LET r_msg = r_msg.subString(1,r_msg.getLength()-1)   
   RETURN r_msg
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2015/8/27 By 04010
# Modify.........: #150803-00008#2 add
################################################################################
PUBLIC FUNCTION cl_xg_maillist(p_gzgh)
   DEFINE p_gzgh		     RECORD LIKE gzgh_t.*
   #190408-00033#3 mark(s)
#   DEFINE l_maillist        DYNAMIC ARRAY OF RECORD   
#              oojn005      LIKE oojn_t.oojn005,  #1:收件者 2:副本 3:密件副本
#              oojn011      LIKE oojn_t.oojn011,  #聯絡對象類型                 
#              oojn003      LIKE oojn_t.oojn003,  #收件人員工編號   
#              oojn004      LIKE oojn_t.oojn004,  #姓名 
#              oojn006      LIKE oojn_t.oojn006,  #Mail address
#              oojn012      LIKE oojn_t.oojn012,  #進階設定   #190107-00075#8 add              
#              oojn007      LIKE oojn_t.oojn007   #報表mail檔案格式               
#                           END RECORD   
   #190408-00033#3 mark(e)
   #190408-00033#3 add(s)
   DEFINE l_maillist        DYNAMIC ARRAY OF RECORD
              oojn005      LIKE gzpe_t.gzpe007,  #1:收件者 2:副本 3:密件副本
              oojn011      LIKE gzpe_t.gzpe012,  #聯絡對象類型
              oojn003      LIKE gzpe_t.gzpe004,  #收件人員工編號 
              oojn004      LIKE gzpe_t.gzpe008,  #姓名 
              oojn006      LIKE gzpe_t.gzpe009,  #Mail address
              oojn012      LIKE gzpe_t.gzpe013,  #進階設定
              oojn007      LIKE gzpe_t.gzpe010   #報表mail檔案格式   
                           END RECORD
   #190408-00033#3 add(e)
   DEFINE l_i              LIKE type_t.num5
   DEFINE li_idx           LIKE type_t.num5
   DEFINE l_subject        VARCHAR(300)            
   DEFINE l_source         STRING
   DEFINE l_str            STRING   

   LET p_gzgh.gzgh030 = NULL
   LET p_gzgh.gzgh031 = NULL
   LET p_gzgh.gzgh032 = NULL
   LET p_gzgh.gzgh052 = NULL   #190107-00075#8 add
   
   IF g_schedule_d.getLength() > 0 THEN
      #230109-00023#1 add(s)
      CALL cl_rpt_get_default_maillist("","","") RETURNING l_subject,l_source,l_maillist   #150901-00021#1 add
    
      IF NOT cl_null(l_subject) AND l_subject <> 'default' THEN LET p_gzgh.gzgh051 = l_subject END IF   
      #240130-00011#2 add(s) 搭配aoos010的個資訊息設定
      LET l_str = cl_get_para(g_enterprise, g_site, "E-SYS-2024")
      IF NOT cl_null(l_str) THEN
         IF NOT cl_null(l_source) AND l_source <> 'default' THEN
            LET p_gzgh.gzgh033 = l_str,'|',l_source
         ELSE
            LET p_gzgh.gzgh033 = l_str,'|','default'
         END IF
      ELSE
         IF NOT cl_null(l_source) AND l_source <> 'default' THEN LET p_gzgh.gzgh033 = l_source END IF      
      END IF
      #IF NOT cl_null(l_source) AND l_source <> 'default' THEN LET p_gzgh.gzgh033 = l_source END IF
      #240130-00011#2 add(e)
      #230109-00023#1 add(e)
      FOR li_idx = 1 TO g_schedule_d.getLength()
#         LET g_schedule_d[li_idx].gzpe009 = s_aooi350_get_user_contact(g_schedule_d[li_idx].gzpe004,'4')   #150803-00008#2 mark
#         LET g_schedule_d[li_idx].gzpe009 = s_aooi350_get_cont_ppl_email(g_schedule_d[li_idx].gzpe012,g_schedule_d[li_idx].gzpe004)   #150803-00008#2 add   #161222-00006#1 mark
         LET g_schedule_d[li_idx].gzpe009 = cl_s_aooi350_get_cont_ppl_email(g_schedule_d[li_idx].gzpe012,g_schedule_d[li_idx].gzpe004)                       #161222-00006#1 add
         #170911-00002#4 add(s)
         #匯出excel加入「只限資料」與「表頭表尾」選項
         IF g_schedule_d[li_idx].gzpe010 = "3" THEN
            IF g_schedule_d[li_idx].l_dataonly = "Y" THEN
               IF g_schedule_d[li_idx].l_HF = "Y" THEN
                  LET g_schedule_d[li_idx].gzpe010 = "9"   #XLS+只限資料+表頭
               ELSE
                  LET g_schedule_d[li_idx].gzpe010 = "7"   #XLS+只限資料
               END IF
            END IF
         END IF
         IF g_schedule_d[li_idx].gzpe010 = "4" THEN
            IF g_schedule_d[li_idx].l_dataonly = "Y" THEN
               IF g_schedule_d[li_idx].l_HF = "Y" THEN
                  LET g_schedule_d[li_idx].gzpe010 = "0"   #XLSX+只限資料+表頭
               ELSE
                  LET g_schedule_d[li_idx].gzpe010 = "8"   #XLSX+只限資料
               END IF
            END IF
         END IF         
         #170911-00002#4 add(e)
         CASE g_schedule_d[li_idx].gzpe007
            WHEN "1"
               LET p_gzgh.gzgh030 = p_gzgh.gzgh030,";",g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010
            WHEN "2"
               LET p_gzgh.gzgh031 = p_gzgh.gzgh031,";",g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010
            WHEN "3"
               LET p_gzgh.gzgh032 = p_gzgh.gzgh032,";",g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010     
            #190107-00075#8 add(s)
            WHEN "4"
               LET p_gzgh.gzgh052 = p_gzgh.gzgh052,"|",g_schedule_d[li_idx].gzpe013,"&",g_schedule_d[li_idx].gzpe010 
#               LET p_gzgh.gzgh052 = p_gzgh.gzgh052,"|::",g_schedule_d[li_idx].gzpe010,":",g_schedule_d[li_idx].gzpe013
            #190107-00075#8 add(e)
         END CASE
      END FOR
   ELSE
      #取得報表預設郵件清單
#      CALL cl_gr_get_default_maillist("","","") RETURNING l_subject,l_source,l_maillist   #150901-00021#1 mark
      CALL cl_rpt_get_default_maillist("","","") RETURNING l_subject,l_source,l_maillist   #150901-00021#1 add
    
      IF NOT cl_null(l_subject) AND l_subject <> 'default' THEN LET p_gzgh.gzgh051 = l_subject END IF   
     #240130-00011#2 add(s) 搭配aoos010的個資訊息設定
      LET l_str = cl_get_para(g_enterprise, g_site, "E-SYS-2024")
      IF NOT cl_null(l_str) THEN
         IF NOT cl_null(l_source) AND l_source <> 'default' THEN
            LET p_gzgh.gzgh033 = l_str,'|',l_source
         ELSE
            LET p_gzgh.gzgh033 = l_str,'|','default'
         END IF
      ELSE
         IF NOT cl_null(l_source) AND l_source <> 'default' THEN LET p_gzgh.gzgh033 = l_source END IF  
      END IF         
      #IF NOT cl_null(l_source) AND l_source <> 'default' THEN LET p_gzgh.gzgh033 = l_source END IF
      #240130-00011#2 add(e)  
      
      FOR l_i = 1 TO l_maillist.getLength()
#         LET l_maillist[l_i].oojn006 = s_aooi350_get_cont_ppl_email(l_maillist[l_i].oojn011,l_maillist[l_i].oojn003) #mail   #161222-00006#1 mark
         LET l_maillist[l_i].oojn006 = cl_s_aooi350_get_cont_ppl_email(l_maillist[l_i].oojn011,l_maillist[l_i].oojn003) #mail #161222-00006#1 add
         
         #轉換預設格式
#         LET l_maillist[l_i].oojn007 = cl_gr_mail_format('X',l_maillist[l_i].oojn007)   #150901-00021#1 mark
         LET l_maillist[l_i].oojn007 = cl_rpt_mail_format('X',l_maillist[l_i].oojn007)   #150901-00021#1 add
         
         CASE l_maillist[l_i].oojn005
            WHEN "1"
               LET p_gzgh.gzgh030 = p_gzgh.gzgh030,";",l_maillist[l_i].oojn006,":",l_maillist[l_i].oojn004,":",l_maillist[l_i].oojn007           
            WHEN "2"
               LET p_gzgh.gzgh031 = p_gzgh.gzgh031,";",l_maillist[l_i].oojn006,":",l_maillist[l_i].oojn004,":",l_maillist[l_i].oojn007           
            WHEN "3"
               #LET p_gzgh.gzgh032 = p_gzgh.gzgh031,";",l_maillist[l_i].oojn006,":",l_maillist[l_i].oojn004,":",l_maillist[l_i].oojn007   #230109-00023#1 mark
               LET p_gzgh.gzgh032 = p_gzgh.gzgh032,";",l_maillist[l_i].oojn006,":",l_maillist[l_i].oojn004,":",l_maillist[l_i].oojn007    #230109-00023#1 add              
            #190107-00075#8 add(s)
            WHEN "4"
               LET p_gzgh.gzgh052 = p_gzgh.gzgh052,"|",l_maillist[l_i].oojn012,"&",l_maillist[l_i].oojn007 
#               LET p_gzgh.gzgh052 = p_gzgh.gzgh052,"|::",l_maillist[l_i].oojn007,":",l_maillist[l_i].oojn012              
            #190107-00075#8 add(e)
               
         END CASE
      END FOR   
   END IF
   
   LET l_str = p_gzgh.gzgh030 CLIPPED
   LET l_str = l_str.trim()
   LET p_gzgh.gzgh030 = l_str.substring(2,LENGTH(l_str))
   
   LET l_str = p_gzgh.gzgh031 CLIPPED
   LET l_str = l_str.trim()        
   LET p_gzgh.gzgh031 = l_str.substring(2,LENGTH(l_str)) 
   
   LET l_str = p_gzgh.gzgh032 CLIPPED
   LET l_str = l_str.trim()                  
   LET p_gzgh.gzgh032 = l_str.substring(2,LENGTH(l_str))   

   #190107-00075#8 add(s)
   LET l_str = p_gzgh.gzgh052 CLIPPED
   LET l_str = l_str.trim()                  
   LET p_gzgh.gzgh052 = l_str.substring(2,LENGTH(l_str)) 
   #190107-00075#8 add(e)
   
   RETURN p_gzgh.*
END FUNCTION

################################################################################
# Descriptions...: 更新收件人mail
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2015/8/27 By 04010
# Modify.........: #150803-00008#4 add
################################################################################
PUBLIC FUNCTION cl_xg_update_gzpe009()
   DEFINE li_cnt       LIKE type_t.num5
   DEFINE l_i          LIKE type_t.num5
   DEFINE l_sql        STRING
   DEFINE l_gzpe011    LIKE gzpe_t.gzpe011   
   DEFINE l_gzpe012    LIKE gzpe_t.gzpe012  
   DEFINE l_gzpe004    LIKE gzpe_t.gzpe004 
   DEFINE l_gzpe009    LIKE gzpe_t.gzpe009    

   IF NOT cl_null(g_schedule_id) THEN
      IF cl_null(g_schedule_seq) OR g_schedule_seq < 0 THEN LET g_schedule_seq = 1 END IF
      
      LET l_sql = "SELECT gzpe011,gzpe012,gzpe004 FROM gzpe_t WHERE gzpeent ='",g_enterprise CLIPPED,"' ",
                  " AND gzpe001 ='",g_schedule_id CLIPPED,"' ",
                  " AND gzpe002 =",g_schedule_seq CLIPPED," AND gzpe003 = 'Z'"
      PREPARE cl_xg_update_gzpe009_pre FROM l_sql
      DECLARE cl_xg_update_gzpe009_cs CURSOR FOR cl_xg_update_gzpe009_pre   
      
      LET l_sql = "SELECT COUNT(1) FROM (",l_sql,") tsed"
      PREPARE cl_xg_update_gzpe009_cnt_pre FROM l_sql
      EXECUTE cl_xg_update_gzpe009_cnt_pre INTO li_cnt
         
      IF li_cnt > 0 THEN
         FOREACH cl_xg_update_gzpe009_cs INTO l_gzpe011,l_gzpe012,l_gzpe004
#            LET l_gzpe009 = s_aooi350_get_cont_ppl_email(l_gzpe012,l_gzpe004)   #161222-00006#1 mark
            LET l_gzpe009 = cl_s_aooi350_get_cont_ppl_email(l_gzpe012,l_gzpe004) #161222-00006#1 add
            UPDATE gzpe_t SET gzpe009 = l_gzpe009
             WHERE gzpeent = g_enterprise AND gzpe001 = g_schedule_id
               AND gzpe002 = g_schedule_seq AND gzpe003 = 'Z'
               AND gzpe011 = l_gzpe011                     
         END FOREACH  
      END IF
   END IF
END FUNCTION

################################################################################
# Descriptions...: 判斷是否為內網IP
# Memo...........: 170209-00022#1 add
# Usage..........: CALL cl_xg_chk_internal_ip(p_ip)
#                  RETURNING r_flag
# Input parameter: p_ip           IP位址
#                : 传入参数变量2   传入参数变量说明2
# Return code....: r_flag         Y:內網;N:外網
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/2/9 By 04010
# Modify.........: #170309-00023#1 停用
################################################################################
PUBLIC FUNCTION cl_xg_chk_internal_ip(p_ip)
   DEFINE p_ip           LIKE type_t.chr20
   DEFINE r_flag         BOOLEAN
   DEFINE l_flag         LIKE type_t.chr1
   DEFINE lst_pk         base.StringTokenizer
   DEFINE ld_ip          DYNAMIC ARRAY OF STRING
   DEFINE li_ip1         LIKE type_t.num5
   DEFINE li_ip2         LIKE type_t.num5

   LET r_flag = FALSE

   #檢查是否為合法IP，虛擬IP有時會為空，不合法即視為內網
   IF NOT cl_xg_chk_validip(p_ip) THEN
      LET r_flag = TRUE
      RETURN r_flag
   END IF

   #將IP拆成陣列
   LET lst_pk = base.StringTokenizer.create(p_ip CLIPPED, ".")
   CALL ld_ip.clear()
   WHILE lst_pk.hasMoreTokens()
      CALL ld_ip.appendElement()
      LET ld_ip[ld_ip.getLength()] = lst_pk.nextToken()
   END WHILE
   
   LET li_ip1 = ld_ip[1]
   #CLASS A(10.0.0.0-10.255.255.255)判斷
   IF li_ip1 = 10 THEN
      LET r_flag = TRUE
      RETURN r_flag
   END IF
   
   #CLASS B(172.16.0.0-172.31.255.255)判斷
   IF li_ip1 = 172 THEN
      LET li_ip2 = ld_ip[2]
      IF li_ip2 < 16 OR li_ip2 > 31 THEN
#         LET r_flag = TRUE   #170309-00023#1 mark
         LET r_flag = FALSE   #170309-00023#1 add
         RETURN r_flag      
      ELSE
#         LET r_flag = FALSE  #170309-00023#1 mark
         LET r_flag = TRUE    #170309-00023#1 add
         RETURN r_flag      
      END IF 
   END IF

   #CLASS C(192.168.0.0-192.168.255.255)判斷
   IF li_ip1 = 192 THEN
      LET li_ip2 = ld_ip[2]
      IF li_ip2 = 168 THEN
         LET r_flag = TRUE
         RETURN r_flag      
      ELSE
         LET r_flag = FALSE
         RETURN r_flag      
      END IF 
   END IF

   RETURN r_flag
END FUNCTION

################################################################################
# Descriptions...: 判斷是否為有效IP
# Memo...........: 170209-00022#1 add
# Usage..........: CALL cl_xg_chk_validip(p_ip)
#                  RETURNING r_chk
# Input parameter: p_ip           传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: r_chk          回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/2/9 By 04010
# Modify.........: #170309-00023#1 停用
################################################################################
PUBLIC FUNCTION cl_xg_chk_validip(p_ip)
   DEFINE p_ip           LIKE type_t.chr20
   DEFINE r_chk          BOOLEAN
   DEFINE ld_array       DYNAMIC ARRAY OF STRING
   DEFINE lst_pk         base.StringTokenizer
   DEFINE l_i            LIKE type_t.num5
   DEFINE l_num          LIKE type_t.num5

   LET r_chk = TRUE
   
   LET lst_pk = base.StringTokenizer.create(p_ip CLIPPED, ".")
   CALL ld_array.clear()
   WHILE lst_pk.hasMoreTokens()
      CALL ld_array.appendElement()
      LET ld_array[ld_array.getLength()] = lst_pk.nextToken()
   END WHILE

   #檢查IP需為X.X.X.X四組數字
   IF ld_array.getLength() <= 3 THEN
      LET r_chk = FALSE
      RETURN r_chk
   END IF

   #僅判別是否落在0-255
   FOR l_i = 1 TO ld_array.getLength()
      LET l_num = ld_array[l_i]
      IF l_num < 0 OR l_num > 255 THEN
         LET r_chk = FALSE
      END IF
   END FOR
   
   RETURN r_chk
END FUNCTION

################################################################################
# Descriptions...: 檢查該作業是否設定『試算表模板設定工具(azzi350)』
# Memo...........: #170301-00053#3 add
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/3/20 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_chk_excel()
   DEFINE l_gzgf000  LIKE gzgf_t.gzgf000   #報表樣板ID
   DEFINE l_gzde008  LIKE gzde_t.gzde008   #客製否
   DEFINE l_sql      STRING
   DEFINE l_cnt      LIKE type_t.num5
   DEFINE r_flag     BOOLEAN
   DEFINE l_roles    STRING  #200428-00006#1 add 判斷相關權限
   
   LET r_flag = FALSE
   
   #190828-00035 add(s)
   #設定試算表串接，但可視需求選擇使用試算表或XG
   IF g_template_flag = 'N' THEN
      RETURN FALSE
   END IF
   #190828-00035 add(e)
   
   #200428-00006#1  --strat--
   LET l_roles = cl_replace_str(g_roles,",","','")
   LET l_roles = "('",l_roles,"')"
   #200428-00006#1  --end--
      
   #判斷是否設定試算表模板
#   LET l_sql = "SELECT DISTINCT gzim013 FROM gzim_t WHERE gzim001=? AND gzim014=? ",
   LET l_sql = "SELECT DISTINCT gzim013 FROM gzim_t WHERE gzim001=? ",
               #"   AND gzim009 = 'default' AND gzim010 = 'default' AND gzimstus = 'Y' AND gzim012 = 'Y' "   #190828-00035#1 add  #200428-00006#1 mark
               "    AND ((gzim009='default' AND gzim010='default')", #200428-00006#1 add
               "    OR   (gzim010 IN ", l_roles CLIPPED," AND gzim009='default')", #顯示符合角色的樣板 #200428-00006#1 add
               "    OR   (gzim010='default' AND gzim009='",g_user CLIPPED,"'))",   #顯示符合用戶的樣板 #200428-00006#1 add
               "    AND gzimstus = 'Y' AND gzim012 = 'Y' " #200428-00006#1 add
               #"   AND gzim009 = 'default' AND gzim010 = 'default' AND gzimstus = 'Y' "                    #190828-00035#1 mark
   PREPARE cl_xg_gzim_pre FROM l_sql
   
   #有客製資料先撈客製資料
#   EXECUTE cl_xg_gzim_pre INTO l_gzgf000 USING g_prog,'c'
#   IF cl_null(l_gzgf000) THEN
#      EXECUTE cl_xg_gzim_pre INTO l_gzgf000 USING g_prog,'s'   
#   END IF

   EXECUTE cl_xg_gzim_pre INTO l_gzgf000 USING g_prog
   
   IF cl_null(l_gzgf000) THEN
      #未設定串接
      LET r_flag = FALSE
      RETURN r_flag
   ELSE
      #有設定串接,需勾稽azzi300與azzi901的客製碼,若報表元件已客製,只撈取客製的azzi300串接的azzi350
      SELECT COUNT(1) INTO l_cnt FROM gzgf_t
       INNER JOIN gzde_t ON gzgf003 = gzde008 AND gzgf001 = gzde001
       WHERE gzgf000 = l_gzgf000
       
      IF l_cnt > 0 THEN
         LET r_flag = TRUE     
      ELSE
         LET r_flag = FALSE
      END IF
      
   END IF
  
   RETURN r_flag
END FUNCTION

################################################################################
# Descriptions...: 試算表模板清單
# Memo...........: 170301-00053#3 add
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/3/29 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_template_list(p_gzim001)
   DEFINE p_gzim001      LIKE gzim_t.gzim001
   DEFINE l_gzde008      LIKE gzde_t.gzde008
   DEFINE l_sql          STRING
   DEFINE l_cnt          LIKE type_t.num5                      
   DEFINE l_roles        STRING   
   DEFINE l_sucess       BOOLEAN
   DEFINE ls_template    STRING   #190828-00035#1 add

   CALL g_gzim.clear()

   LET l_roles = cl_replace_str(g_roles,",","','")
   LET l_roles = "('",l_roles,"')"   
   
   #報表元件執行需勾稽客製碼
   IF NOT cl_null(g_rep_code) THEN
      SELECT gzde008 INTO l_gzde008 FROM gzde_t
       WHERE gzde001 = g_rep_code AND gzdestus = "Y" 
   END IF
   
   LET l_sql = "SELECT DISTINCT 'N',gziml004,gzim002,gzim014,'','',gzim000,gzim009,gzim010,gzim015,gzim007,'' ",
               " FROM gzim_t LEFT JOIN gziml_t on gzim000 = gziml000 AND gziml003='", g_lang CLIPPED, "' ",
               " INNER JOIN gzgf_t ON gzgf000 = gzim013 AND gzgf003 = '",l_gzde008 CLIPPED,"' ", #客製碼須符合
               " WHERE gzim001 = '",g_prog CLIPPED,"' ",
               " AND gzimstus = 'Y' ",
               " AND ((gzim009='default' AND gzim010='default')",
               " OR   (gzim010 IN ", l_roles CLIPPED," AND gzim009='default')", #顯示符合角色的樣板
               " OR   (gzim010='default' AND gzim009='",g_user CLIPPED,"'))"    #顯示符合用戶的樣板
#              ," GROUP BY gzim000 "
                
   #190828-00035#1 add(s)
   #多個試算表模板時，能指定使用的模板
   LET ls_template = g_template
   IF NOT cl_null(g_template) AND ls_template.getIndexOf("_e",1) > 0 THEN
      LET l_sql = l_sql, " AND gzim002 = '",g_template,"' "
   END IF
   #190828-00035#1 add(e)
   
   DECLARE cl_xg_gzim_cs CURSOR FROM l_sql
   
   LET l_sql = "SELECT COUNT(1) FROM (",l_sql,") tsed"
   PREPARE cl_xg_gzim_cnt_pr FROM l_sql
   EXECUTE cl_xg_gzim_cnt_pr INTO l_cnt
   FREE cl_xg_gzim_cnt_pr

   #樣板筆數超過一筆時須提供開窗
   IF l_cnt > 1 THEN 
      CALL cl_xg_dialog()
   ELSE
      EXECUTE cl_xg_gzim_cs INTO g_gzim[1].*
      CALL cl_spreadsheet_xg_fill(g_gzim[1].gzim000,g_rep_tmpname[1])
   END IF


END FUNCTION

################################################################################
# Descriptions...: 樣板選擇
# Memo...........: 170301-00053#3 add
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/3/31 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_dialog()
   DEFINE l_ac              LIKE type_t.num5
   DEFINE l_count           INTEGER
   DEFINE l_dialog          ui.Dialog
   DEFINE l_sucess          BOOLEAN

   OPEN WINDOW w_cl_xg_s01 WITH FORM cl_ap_formpath("lib","cl_xg_s01")         
      ATTRIBUTES(STYLE="layoutwin")
   CALL cl_ui_init()

   CALL cl_xg_gzim_b_fill()
   
   
            DIALOG ATTRIBUTE(UNBUFFERED)
   
               DISPLAY ARRAY g_gzim TO s_detail1.* ATTRIBUTE(COUNT = g_rec_b1)
              
                  BEFORE DISPLAY
                     LET l_count = 1
                     LET g_gzim[1].pick = 'Y'                 #初始狀態第一行必勾選
                     FOR l_count = 2 TO g_gzim.getLength()    #其餘筆數均預設不勾選
                        LET g_gzim[l_count].pick = 'N'
                     END FOR
                     
                        CALL g_gzim.deleteElement(l_count) #刪除系統自動產生的最後一筆空行
                        LET l_ac = 1

                     
                  BEFORE ROW
                     IF g_gzim.getLength() != 1 AND l_ac != 0 THEN          #若不只一筆樣板檔
                        IF l_ac != l_dialog.getCurrentRow("s_detail1") THEN #選擇新筆資料時
                           LET g_gzim[l_ac].pick = 'N'                      #將上筆取消勾選
                           LET l_ac = l_dialog.getCurrentRow("s_detail1")   #取得新筆行數
                           LET g_gzim[l_ac].pick = 'Y'                      #新筆設置為勾選
                        END IF
                     END IF
    
                  AFTER ROW
                     LET l_ac = l_dialog.getCurrentRow("s_detail1")
                     
              END DISPLAY   
              
              BEFORE DIALOG
                 LET l_dialog = ui.Dialog.getCurrent()
                 
              ON ACTION btn_accept
                  ACCEPT DIALOG
   
              ON ACTION btn_cancel
                  LET g_rep_flag = TRUE
                  EXIT DIALOG
   
              ON ACTION CLOSE
                  LET g_rep_flag = TRUE
                  EXIT DIALOG   
            END DIALOG
            
   IF NOT g_rep_flag THEN
      CALL cl_spreadsheet_xg_fill(g_gzim[l_ac].gzim000,g_rep_tmpname[1])
   END IF
            
   #畫面關閉
   CLOSE WINDOW w_cl_xg_s01
   
   RETURN
END FUNCTION

################################################################################
# Descriptions...: 試算表模板清單
# Memo...........: 170301-00053#3 add
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/3/31 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_gzim_b_fill()
   DEFINE li_cnt    LIKE type_t.num10

   LET li_cnt = 1
   FOREACH cl_xg_gzim_cs INTO g_gzim[li_cnt].*
      IF SQLCA.sqlcode THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = SQLCA.sqlcode
         LET g_errparam.extend = 'FOREACH:'
         LET g_errparam.popup = TRUE
         CALL cl_err()

         EXIT FOREACH
      END IF
      IF cl_null(g_gzim[li_cnt].gzim015) THEN LET g_gzim[li_cnt].gzim015 = 'N' END IF
      
      #使用者名稱
      IF g_gzim[li_cnt].gzim009 = "default" THEN
         LET g_gzim[li_cnt].gzim009_desc = cl_getmsg("azz-00665",g_lang)
      ELSE
         INITIALIZE g_ref_fields TO NULL
         LET g_ref_fields[1] = g_gzim[li_cnt].gzim009
         CALL ap_ref_array2(g_ref_fields,"SELECT ooag011 FROM ooag_t WHERE ooagent='"||g_enterprise||"' AND ooag001=? ","") RETURNING g_rtn_fields
         LET g_gzim[li_cnt].gzim009_desc = '', g_rtn_fields[1] , ''
      END IF
      
      #角色名稱
      IF g_gzim[li_cnt].gzim010 = "default" THEN
         LET g_gzim[li_cnt].gzim010_desc = cl_getmsg("azz-00666",g_lang)
      ELSE
         INITIALIZE g_ref_fields TO NULL
         LET g_ref_fields[1] = g_gzim[li_cnt].gzim010
         CALL ap_ref_array2(g_ref_fields,"SELECT gzyal003 FROM gzyal_t WHERE gzyalent='"||g_enterprise||"' AND gzyal001=? AND gzyal002='"||g_dlang||"'","") RETURNING g_rtn_fields
         LET g_gzim[li_cnt].gzim010_desc = '', g_rtn_fields[1] , ''
      END IF
      
      
      IF li_cnt = 1 THEN LET g_gzim[li_cnt].pick ='Y' END  IF  
      LET li_cnt = li_cnt + 1
   END FOREACH

   CLOSE cl_xg_gzim_cs
   FREE cl_xg_gzim_cs
   CALL g_gzim.deleteElement(li_cnt)
   LET g_rec_b1 = li_cnt - 1
   
   #依優先序重新排序陣列內容
   CALL cl_xg_b_fill_sort()
   
END FUNCTION

################################################################################
# Descriptions...: 依優先序重新排序陣列內容
# Memo...........: 170301-00053#3 add
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/3/31 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_b_fill_sort()
   DEFINE l_i,l_j,l_max  LIKE type_t.num10
   DEFINE l_gzim_cnt     LIKE type_t.num10
   DEFINE l_gzim_tmp     DYNAMIC ARRAY OF  type_g_gzim
   DEFINE l_gzim_t       type_g_gzim

   IF g_gzim.getlength() = 0 THEN RETURN END IF

   #先跑一圈，賦予每一行的樣板排序碼
   LET l_max = 0 
   FOR l_i = 1 TO g_gzim.getlength()
      IF cl_null(g_gzim[l_i].gzim009) THEN CONTINUE FOR END IF
      LET l_gzim_tmp[l_i].sort_num = 0

      #用戶
      IF g_gzim[l_i].gzim009 = g_user THEN
         IF g_gzim[l_i].gzim015 = "Y" THEN 
            LET l_gzim_tmp[l_i].sort_num = l_gzim_tmp[l_i].sort_num + 1
         ELSE
            LET l_gzim_tmp[l_i].sort_num = l_gzim_tmp[l_i].sort_num + 2
         END IF           
      END IF

      #角色
      IF g_roles.getIndexOf(g_gzim[l_i].gzim010 CLIPPED,1) > 0 THEN
         IF g_gzim[l_i].gzim015 = "Y" THEN 
            LET l_gzim_tmp[l_i].sort_num = l_gzim_tmp[l_i].sort_num + 4 
         ELSE
            LET l_gzim_tmp[l_i].sort_num = l_gzim_tmp[l_i].sort_num + 8
         END IF          
      END IF

      #預設樣版
      IF g_gzim[l_i].gzim015 = 'Y' THEN
         LET l_gzim_tmp[l_i].sort_num =l_gzim_tmp[l_i].sort_num + 16 
      ELSE
         LET l_gzim_tmp[l_i].sort_num =l_gzim_tmp[l_i].sort_num + 32
      END IF

      IF l_gzim_tmp[l_i].sort_num > l_max THEN
         LET l_max = l_gzim_tmp[l_i].sort_num 
      END IF
   END FOR

   LET l_gzim_cnt = l_gzim_tmp.getlength()
   #用選擇排序法
   FOR l_i = 1 TO l_gzim_cnt
      FOR l_j = l_i+1 TO l_gzim_cnt        
         #IF l_gzim_tmp[l_j].sort_num > l_gzim_tmp[l_i].sort_num THEN   #170622-00040#1 mark
         IF l_gzim_tmp[l_j].sort_num < l_gzim_tmp[l_i].sort_num THEN   #170622-00040#1 add
            #資料互換
            INITIALIZE l_gzim_t.* TO NULL
            LET l_gzim_t.gziml004 = g_gzim[l_i].gziml004
            LET l_gzim_t.gzim002  = g_gzim[l_i].gzim002 
            LET l_gzim_t.gzim014  = g_gzim[l_i].gzim014
            LET l_gzim_t.gzim009_desc  = g_gzim[l_i].gzim009_desc
            LET l_gzim_t.gzim010_desc  = g_gzim[l_i].gzim010_desc            
            LET l_gzim_t.gzim000  = g_gzim[l_i].gzim000
            LET l_gzim_t.gzim009  = g_gzim[l_i].gzim009
            LET l_gzim_t.gzim010  = g_gzim[l_i].gzim010
            LET l_gzim_t.gzim015  = g_gzim[l_i].gzim015
            LET l_gzim_t.gzim007  = g_gzim[l_i].gzim007
            LET l_gzim_t.sort_num = l_gzim_tmp[l_i].sort_num
                
            LET g_gzim[l_i].gziml004 = g_gzim[l_j].gziml004
            LET g_gzim[l_i].gzim002  = g_gzim[l_j].gzim002 
            LET g_gzim[l_i].gzim014  = g_gzim[l_j].gzim014
            LET g_gzim[l_i].gzim009_desc  = g_gzim[l_j].gzim009_desc
            LET g_gzim[l_i].gzim010_desc  = g_gzim[l_j].gzim010_desc
            LET g_gzim[l_i].gzim000  = g_gzim[l_j].gzim000
            LET g_gzim[l_i].gzim009  = g_gzim[l_j].gzim009
            LET g_gzim[l_i].gzim010  = g_gzim[l_j].gzim010
            LET g_gzim[l_i].gzim015  = g_gzim[l_j].gzim015
            LET g_gzim[l_i].gzim007  = g_gzim[l_j].gzim007
            LET l_gzim_tmp[l_i].sort_num = l_gzim_tmp[l_j].sort_num                                       

            LET g_gzim[l_j].gziml004 = l_gzim_t.gziml004
            LET g_gzim[l_j].gzim002  = l_gzim_t.gzim002 
            LET g_gzim[l_j].gzim014  = l_gzim_t.gzim014 
            LET g_gzim[l_j].gzim009_desc  = l_gzim_t.gzim009_desc 
            LET g_gzim[l_j].gzim010_desc  = l_gzim_t.gzim010_desc 
            LET g_gzim[l_j].gzim000  = l_gzim_t.gzim000 
            LET g_gzim[l_j].gzim009  = l_gzim_t.gzim009 
            LET g_gzim[l_j].gzim010  = l_gzim_t.gzim010 
            LET g_gzim[l_j].gzim015  = l_gzim_t.gzim015 
            LET g_gzim[l_j].gzim007  = l_gzim_t.gzim007 
            LET l_gzim_tmp[l_j].sort_num = l_gzim_t.sort_num
         END IF
      END FOR
   END FOR
    
   FOR l_i = 1 TO l_gzim_tmp.getlength()
      IF l_i=1 THEN
         LET g_gzim[l_i].pick = 'Y'
      ELSE
         LET g_gzim[l_i].pick = 'N'
      END IF
   END FOR

END FUNCTION

################################################################################
# Descriptions...: 檢查暫存表欄位型態不一致
# Memo...........: 170619-00003#1 add
# Usage..........: CALL cl_xg_chk_tmp_table_field_type(p_n,p_gzgf000)
#                  RETURNING r_msg
# Input parameter: p_n            暫存表序號
#                : p_gzgf000      樣板ID
# Return code....: r_msg          欄位清單
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/6/19 By BingYang
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_chk_tmp_table_field_type(p_n,p_gzgf000)
    DEFINE p_n          LIKE type_t.num10
    DEFINE p_gzgf000    LIKE gzgf_t.gzgf000
    DEFINE l_sql        STRING
    DEFINE l_columnname LIKE gzgg_t.gzgg001
    DEFINE l_datatype   VARCHAR(15)
    DEFINE ld_gzgg1     DYNAMIC ARRAY OF RECORD
        gzgg001 LIKE gzgg_t.gzgg001,
        gzgg003 LIKE gzgg_t.gzgg003
        END RECORD
    DEFINE ld_gzgg2     DYNAMIC ARRAY OF RECORD
        gzgg001 LIKE gzgg_t.gzgg001,
        gzgg003 LIKE gzgg_t.gzgg003
        END RECORD
    DEFINE l_i          LIKE type_t.num5
    DEFINE r_msg        STRING
    DEFINE l_report     STRING   #240130-00011#2 add
    DEFINE l_j          LIKE type_t.num5 #240612-00039#3 add     
    
    LET r_msg = NULL
    LET l_report = g_report   #240130-00011#2 add
    
    #PGS(D)-05292
    #240130-00011#2 mark(s)
    #LET l_sql = "SELECT LOWER(COLUMN_NAME), LOWER(DATA_TYPE) FROM ALL_TAB_COLUMNS 
    #    WHERE LOWER(TABLE_NAME) = '", g_rep_tmpname[p_n] CLIPPED, "' 
    #    AND LOWER(OWNER) = '", g_report CLIPPED, "' ORDER BY LOWER(COLUMN_NAME)"
    #240130-00011#2 mark(e)
    #240130-00011#2 add(s)
    LET l_sql = "   SELECT LOWER(COLUMN_NAME), ",
                "          LOWER(DATA_TYPE) ",
                "     FROM ALL_TAB_COLUMNS ", 
                "    WHERE     TABLE_NAME = '", g_rep_tmpname[p_n].toUpperCase() CLIPPED, "' ", 
                "          AND OWNER = '", l_report.toUpperCase() CLIPPED, "' ",
                " ORDER BY LOWER(COLUMN_NAME)"        #240329-00013#1 add
                #" ORDER BY COLUMN_NAME"              #240329-00013#1 mark
    #240130-00011#2 add(e)
    #PGS(D)-05292
    DECLARE cl_xg_chk_field1 CURSOR FROM l_sql
    
    CALL ld_gzgg1.clear()
    FOREACH cl_xg_chk_field1 INTO l_columnname, l_datatype
        CALL ld_gzgg1.appendElement()
        LET ld_gzgg1[ld_gzgg1.getLength()].gzgg001 = l_columnname
        LET l_datatype = l_datatype
        CASE
            WHEN l_datatype = "varchar" OR l_datatype = "varchar2"
                LET ld_gzgg1[ld_gzgg1.getLength()].gzgg003 = 'C'
            WHEN l_datatype = "number" OR l_datatype = "integer" OR l_datatype = "smallint" #231017-00022 add integer、smallint  
                                       OR l_datatype = "numeric" OR l_datatype = "int4" OR l_datatype = "int2"    #240612-00039#3 add numeric、int4、int2
                LET ld_gzgg1[ld_gzgg1.getLength()].gzgg003 = 'N'
            WHEN l_datatype = "date" OR l_datatype = "datetime"
                LET ld_gzgg1[ld_gzgg1.getLength()].gzgg003 = 'D'
            #PGS(D)-05292
            #PGS(D)-05929
            OTHERWISE
                LET ld_gzgg1[ld_gzgg1.getLength()].gzgg003 = 'C'
        END CASE
    END FOREACH
    
    IF p_n = 1 THEN     #主報表
        LET l_sql = "SELECT gzgg001, gzgg003 FROM gzgg_t 
            WHERE gzgg000 = '", p_gzgf000, "' AND gzgg002 = '", g_rlang, "' 
            AND (gzgg017 IS NULL OR gzgg017 = 'Y' OR gzgg017 = 'N') ORDER BY LOWER(gzgg001)"
    ELSE
        LET l_sql = "SELECT gzgg001, gzgg003 FROM gzgg_t 
            WHERE gzgg000 = '", p_gzgf000, "' AND gzgg002 = '", g_rlang, "' 
            AND gzgg017 = '", p_n CLIPPED, "' ORDER BY LOWER(gzgg001)"
    END IF
    DECLARE cl_xg_chk_field2 CURSOR FROM l_sql
    
    LET l_i = 1
    CALL ld_gzgg2.clear()
    FOREACH cl_xg_chk_field2 INTO ld_gzgg2[l_i].gzgg001, ld_gzgg2[l_i].gzgg003
        LET l_i = l_i + 1
    END FOREACH
    
    #PGS(D)-05292
    #240612-00039#3 modify(s) 調整先判斷欄位名稱是否相同
    FOR l_i = 1 TO ld_gzgg1.getLength()
        FOR l_j = 1 TO ld_gzgg2.getLength()
           IF ld_gzgg1[l_i].gzgg001 == ld_gzgg2[l_j].gzgg001 THEN
              IF ld_gzgg1[l_i].gzgg003 <> ld_gzgg2[l_j].gzgg003 THEN
                  IF p_n = 1 THEN
                      IF cl_null(r_msg) THEN
                          LET r_msg = cl_getmsg("lib-00241",g_lang) CLIPPED,ld_gzgg1[l_i].gzgg001,","
                      ELSE
                          LET r_msg = r_msg,ld_gzgg1[l_i].gzgg001,","
                      END IF
                  ELSE
                      IF cl_null(r_msg) THEN
                          LET r_msg = cl_getmsg("lib-00242",g_lang) CLIPPED,ld_gzgg1[l_i].gzgg001,","
                      ELSE
                          LET r_msg = r_msg,ld_gzgg1[l_i].gzgg001,","
                      END IF
                  END IF
              ELSE
                  EXIT FOR
              END IF
          END IF
       END FOR
    END FOR
    #240612-00039#3 modify(e)
    #PGS(D)-05292
    
    LET r_msg = r_msg.subString(1,r_msg.getLength()-1)
    RETURN r_msg
END FUNCTION

################################################################################
# Descriptions...: 取得Client端OS版本
# Memo...........: #180919-00051#1 add
#                  原廠功能在GDC 2.50不work, CALL ui.Interface.frontCall("standard", "feInfo", ["osVersion"], [res])
# Usage..........: CALL cl_xg_osVersion()
#                  RETURNING r_ver
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2018/9/19 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_xg_osVersion()
   DEFINE l_wcpath            STRING   #components路徑
   DEFINE r_ver               STRING
   
   #設定webcomponent路徑,需在開啟視窗前設定
   IF cl_null(FGL_GETENV("GUIMODE")) OR FGL_GETENV("GUIMODE") = "GDC" THEN
      LET l_wcpath = FGL_GETENV("FGLASIP"),"/components"
      CALL ui.interface.frontCall("standard", "setwebcomponentpath",[l_wcpath],[])   #不支援GWC
   END IF
   
   #Fix for GBC add --- start ---
   #OPEN WINDOW cl_xg_s02 WITH FORM cl_ap_formpath("lib","cl_xg_s02") #Fix for GBC mark
   IF FGL_GETENV('T100NANAWEB') = 'Y' THEN
      OPEN WINDOW cl_xg_s02 WITH FORM cl_ap_formpath("lib","cl_xg_s02") ATTRIBUTES(STYLE='layoutwin')
   ELSE
      OPEN WINDOW cl_xg_s02 WITH FORM cl_ap_formpath("lib","cl_xg_s02")
   END IF
   #Fix for GBC add --- end ---   
   
   DIALOG ATTRIBUTES(UNBUFFERED=TRUE)
      INPUT r_ver FROM l_osver ATTRIBUTES(WITHOUT DEFAULTS=TRUE)
         #WebCompnent觸發的4gl Action
         ON ACTION wc_action
            DISPLAY "OS version :", r_ver
            EXIT DIALOG
      END INPUT
   
      ON ACTION close
         EXIT DIALOG
   
   END DIALOG
   
   CLOSE WINDOW cl_xg_s02
   RETURN r_ver

END FUNCTION

 
{</section>}
 
