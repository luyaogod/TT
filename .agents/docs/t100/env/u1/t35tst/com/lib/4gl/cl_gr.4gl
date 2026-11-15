#該程式未解開Section, 採用最新樣板產出!
#該程式為freestyle程式!
{<section id="cl_gr.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0135(1900-01-01 00:00:00), PR版次:0135(2026-04-30 16:59:27)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000837
#+ Filename...: cl_gr
#+ Description: GR library
#+ Creator....: 04010(2014-04-21 09:09:51)
#+ Modifier...: 00000 -SD/PR- 14910
 
{</section>}
 
{<section id="cl_gr.global" type="s" >}
#應用 p00 樣板自動產生(Version:6)
#add-point:填寫註解說明 name="main.memo"
############################################################################################################################
#報表執行流程, 以報表作業axmr500, 報表元件axmr500_g01為例
#1.一般列印:執行axmr500>點擊列印按鈕>將QBE傳給axmr500_g01>呼叫cl_gr_handler>選擇報表樣板與列印格式>載入報表API>產生報表資料>印出報表
############################################################################################################################
#+ Modifier...: No.150526-00012#2 15/05/26 04010   t類串r類GR列印視窗開窗Style
#+ Modifier...: No.150611-00032#3 15/06/12 04010   表尾功能,排程mail address
#+ Modifier...: No.150709-00023#1 15/07/10 04010   1.採購發mail流程需求,調整郵寄報表主旨、內文規則
#                                                  2.因應T類cmdrunｍ串R類,g_prog會變R類,判斷BPM簽核前先將g_prog換成g_parentprog
#                                                  3.信件主旨獨立FUNCTION,與訊息中心共用
#+ Modifier...: No.150803-00008#3 15/08/07 04010   1.報表匯出權限控管
#                                                  2.報表預設郵件清單
#+ Modifier...: No.150901-00021#1 15/09/01 04010   將adz中,使用到cl_gr跟cl_xg的FUNCTION獨立出來, 降低關聯性
#+ Modifier...: No.150904-00005#4 15/09/11 04010   gzgh046加入表尾說明
#+ Modifier...: No.150904-00005#5 15/09/11 04010   列印條件
#+ Modifier...: No.150916-00002#1 15/09/16 02716   環境參數讀取資料庫
#+ Modifier...: No.150921-00060#1 15/09/25 04010   增加系統印表機,快速列印
#+ Modifier...: No.150922-00009#1 15/10/01 05177   報表加密碼
#+ Modifier...: No.150930-00022#3 15/10/13 04010   調整匯出Excel參數,保留空白
#+ Modifier...: No.151021-00013#1 15/10/21 05177   報表加密碼功能調整,內部郵件若取不到用戶密碼,則以郵件帳號加密
#+ Modifier...: No.151027-00003#1 15/10/27 02716   判斷報表相關參數是否設定
#+ Modifier...: No.151201-00015#1 15/10/27 04010   直接送印造成迴圈問題排除
#+ Modifier...: No.160108-00011#1 16/01/08 04010   logo修正
#+ Modifier...: No.160113-00004#1 16/01/13 04010   取得郵件寄件人資訊FUNCTION改成PUBLIC, 與cl_xg共用
#+ Modifier...: No.160323-00014#2 16/03/29 04010   1.呼叫的sub改lib(s_aooi200_get_slip->cl_getslip_get_slip)
#                                                  2.oojk改為gzgp
#+ Modifier...: No.160608-00012#1 16/06/08 04010   1.T100 Light架構調整
#                                                  2.加入sampledata功能
#+ Modifier...: No.160817-00023#1 16/08/17 02716   調整T作業可以直接送印報表
#                                 16/08/24 04010   日期格式使用aooi100的設定,原本固化3種選項,需動態取得
#+ Modifier...: No.160912-00001#1 16/09/12 04010   1.修正Server直接送印時,中一刀列印方向錯誤,原先路徑為報表主機的D:\T100_gr,修正為ERP主機的/T100_gr
#                                                  2.增加log機制
#+ Modifier...: No.160921-00007#3 16/09/21 02716   調整報表作業列印報表程式,報表名稱、排程的抓法以報表作業為主
#+ Modifier...: No.161013-00029#1 16/10/13 02716   調整直接送印若g_template有值，需多判斷gzgn004
#+ Modifier...: No.161103-00014#1 16/11/03 02716   調整T類直接送印若g_template沒有值，需多判斷g_rep_code(報表元件)
#+ Modifier...: No.161128-00002#1 16/11/29 04010   收件者名稱取得方式變更，需搭配161125-00006#1
#+ Modifier...: No.161219-00003#1 16/12/19 04010   BPM簽核修正
#+ Modifier...: No.161214-00061#1 16/12/19 04010   表頭表尾條碼功能
#+ Modifier...: No.161222-00006#1 16/12/22 04010   將需要的最小項目移入 cl_s_aooi350，需搭配161222-00001#1
#+ Modifier...: No.161227-00007#1 16/12/27 02716   調整報表列印標題，報表title名稱等抓取報表元件，代號用報表作業。
#+ Modifier...: No.170104-00023#1 17/01/05 02716   透過背景直接送印多報表元件，故將g_bg_cnt移至top_report.inc
#+ Modifier...: No.170105-00009#2 17/01/05 04010   快速列印種類增加上SVG的選項
#+ Modifier...: No.170120-00031#1 17/01/20 04010   T類列印報表時，在寄送郵件的視窗中，部分按鈕的圖示未顯示
#+ Modifier...: No.161121-00016#1 17/02/09 02716   對內對外名稱因aooi100未設該語言，造成列印報表時，彈出多個訊息視窗
#+ Modifier...: No.170223-00005#2 17/03/01 04010   BPM簽核需帶出簽核日期與簽核時間
#+ Modifier...: No.170316-00056#1 17/03/16 04010   修正直接送印時,已客製的報表元件仍取標準樣板清單的問題,需判斷客製碼
#+ Modifier...: No.170301-00053#2 17/03/30 04010   將cl_gr_chk_tiptopgroup改為PUBLIC
#+ Modifier...: No.170426-00038#1 17/05/03 04010   在LIB中判斷g_xml.subject若有值則使用, 可在報表程式中自行給定郵件主旨
#+ Modifier...: No.170524-00050#1 17/05/24 04010   1.收件人名稱若為空, 收不到附件
#                                                  2.GR郵件開窗,程式中指定的信件內文被清空問題
#+ Modifier...: No.170525-00105#1 17/05/25 04010   SELECT COUNT(*) 變更為 SELECT COUNT(1)
#+ Modifier...: No.170602-00007#1 17/06/02 04010   1.GR報表郵件視窗調整內文後未顯示在郵件中
#                                                  2.郵件本文中的日期改用cl_get_current()
#+ Modifier...: No.170406-00030#1 17/04/07 04010   為避免mount連線斷掉導致報表無法列印,原判斷檔案是否存在與紙張檢查,改檢查AP主機的路徑
#                                                  正確呼叫方式fgl_report_loadCurrentSettings("axmr500_g01.4rp")
#                                                  fgl_report_configureDistributedEnvironment(WINFGLDIR,WINFGLPROFILE,"D:\T100_gr\topprd\axm\4rp\zh_TW\",NULL)
#+ Modifier...: No.170618-00057#1 17/06/18 04010   apmt500列印報表時，在「報表預設郵件清單設定作業(aooi998)」已設定，但開啟郵件視窗並未帶入預設清單
#+ Modifier...: No.170710-00060#1 17/07/11 09760   系統於背景執行報表作業時，可依設定進行Server直接送印
#+ Modifier...: No.170725-00086#1 17/07/26 04010   不使用mount目錄，避免連線不穩導致異常，調整歷史報表與郵件寄送部分
#+ Modifier...: No.170808-00030#1 17/08/08 04010   提供快速下載報表資料(sampledata)，以便開發人員遇到問題時能快速提供所需資料
#+ Modifier...: No.170815-00008#1 17/08/15 04010   1.GR列印視窗sampledata併入下拉選單中 2.壓縮檔案時加入圖片檔 3.日後若要開放用戶設定預設樣板,請將欄位加在上方樣板清單中
#+ Modifier...: No.170828-00002#1 17/08/28 04010   憑證報表Server直接送印，指定紙張方向，避免中一刀報表列印結果異常
#+ Modifier...: No.170912-00018#1 17/09/12 04010   為方便問題排查，將170406-00030#1調整的API參數改回
#+ Modifier...: No.170928-00029#1 17/09/28 04010   預設郵件信件內文未帶入
#+ Modifier...: No.170929-00002#1 17/10/02 04010   1.檢查聯絡對象編號是否輸入錯誤
#                                                  2.調整收件人郵件自動帶入
#+ Modifier...: No.171020-00013#1 17/10/20 04010   郵件視窗, 選擇"交易對象編號",enter後名稱被清空
#+ Modifier...: No.171206-00020#1 17/12/06 04010   憑證報表列印時, 需能指定檔名與列印格式
#+ Modifier...: No.171211-00055#1 17/12/12 04010   匯出Excel列高自動延展, 需搭配GRE 2.50.29, 並調整API
#                                                  將fgl_report_configureXLSDevice()的參數 "ignoreColumnAlignment" and "ignoreRowAlignment" 設成FALSE
#+ Modifier...: No.171211-00031#1 17/12/12 04010   BPM簽核關卡可依設定決定排序(升冪/降冪),需一併更新171211-00039#1
#+ Modifier...: No.170914-00030#1 17/12/27 04010   寄送郵件後彈窗提示成功否，僅為Javamail狀態，無法取得接收方訊息
#+ Modifier...: No.180123-00036#1 18/01/23 04010   新增中越文樣板功能
#+ Modifier...: No.180131-00079#1 18/01/31 04010   中越文樣板,因g_dlang可依客戶需求設定,例如: 高階主管只看英文資料,越文語系登入也需登打英文資料
#                                                  表頭表尾代號應以g_rlang取得,與實體樣板相同語系,才不會有條碼資料問題
#+ Modifier...: No.180213-00002#1 18/02/13 04010   使用g_bg4rp變數, 讓GR取得樣板的彈性更大, ex: axmr500中的英文樣板、中越樣板, LET g_bg4rp = " gzgd002 <> 'axmr500_g01_01' "
#+ Modifier...: No.180402-00004#1 18/04/02 04010   當為中越樣板時,表頭表尾固定取中文+越文說明
#+ Modifier...: N0.180420-00001#4 18/04/20 11084   新增表頭表尾自定義欄位選項
#+ Modifier...: No.180524-00013#2 18/05/29 11084   將預設報表元件帶入YY表頭表尾設定，新增判斷要帶入T100_YY或YY，用意為客戶家可能未更新資料庫，資料還停留在舊的
#+ Modifier...: No.180920-00021#1 18/09/20 04010   附件檔名使用cl_get_current()取得的時間沒有毫秒，若太頻繁發送郵件，會導致檔名重複，無法收到正確附件，改用cl_get_timestamp()取值。
#+ Modifier...: No.180904-00016#1 18/09/20 11084   於azzi330_01新增一個統一編號選項，當p_gzgk006選擇30時需撈取統一編號資訊
#+ Modifier...: No.180920-00065#1 18/09/20 11084   修改歷史報表留存份數FUNCTION SQL段enterprise
#+ Modifier...: No.180719-00037#1 18/09/26 11084   修正排程未依指定樣板寄送報表郵件
#+ Modifier...: No.181123-00035#1 18/11/27 04010   修正T類設定直接送印未依指定樣板呈現
#+ Modifier...: No.181213-00050#1 18/12/19 02716   增加3.10架構的程式段使其並行
#+ Modifier...: No.190117-00039#1 19/01/19 02716   將憑證列印視窗調整為一次性使用，按下確定後即關窗，若用戶勾選保留視窗時，回寫azzi800的個人值
#+ Modifier...: No.190107-00075#6 19/01/19 04010   郵件設定增加「進階設定」，提供郵件合併寄送功能
#+ Modifier...: No.190212-00037#4 19/02/22 11084   新增cl_gr_chk_code FUNCTION，判斷是否顯示條碼
#+ Modifier...: No.190321-00031#1 19/03/21 11084   修正GR列印詢問視窗flag判斷錯誤問題
#                                                  1.選擇sampledata未產生檔案，點擊第二次才產生檔案問題
#                                                  2.郵件寄送後視窗，不保留視窗，視窗依舊存在問題
#                                                  3.當儲存歷史報表狀態，列印視窗若不保留，僅能列印一次問題
#+ Modifier...: No.190327-00044#1 19/04/11 04010   BPM流程串接T100作業，使用GWC模式，1.不允許選擇SVG，但能匯出報表。 2.g_site可能為ALL，需撈取使用者的預設site。
#+ Modifier...: No.190422-00009#1 19/04/26 11084   修正寄送郵件時，若主旨或本文有單引號寄送失敗問題
#+ Modifier...: No.190503-00009#1 19/05/03 11084   DISPLAY出Template資訊，排程時可以抓到更多資訊
#+ Modifier...: No.190223-00010#1 19/05/14 11084   新增當寄送排程時，若無報表資料，則寄送無報表資料通知信件
#+ Modifier...: No.190408-00033#2 19/05/27 04010   1.T100V3.0匯出其它格式需額外呼叫API
#                                                  2.因原先LIKE oojn_t的欄位型態未被SELECT使用，故將LIKE oojn_t型態的欄位調整為LIKE gzpe_t對應的欄位。
#+ Modifier...: No.190606-00010#1 19/06/25 11084   修正郵件寄送問題
#                                                  1.為避免客戶主旨需要使用到一般符號需求，將原先以|組成的Mail寄送資訊，調整用{+}組合符號。
#                                                  2.將原先加密寄送段用echo方式寫入資訊部分，改以Channel方式做寫入，避免遇到190422-00009#1問題。
#                                                  3.修正指定時間排程，無法正常指定報表主旨、內文問題。
#+ Modifier...: No.190624-00029#1 19/06/26 11084   修正開窗郵件寄送後，跳出通知視窗若未按確定，不會接續寄送下一筆郵件問題，調整為將收件者資訊保留，整合後寄完最後一筆後，彈出寄件成功失敗通知資訊。
#+ Modifier...: No.190705-00001#1 19/07/05 11084   修正英文樣板列印時，若未有英文語系scc導致crash問題
#                                                  1.修正呼叫FUNCTION後需要回傳值時，遇到錯誤時需要RETURN未給予NULL導致程式Crash問題
#                                                  2.調整若未對應scc資料，僅POP一次錯誤訊息
#+ Modifier...: No.190710-00007#1 19/08/06 11084   修正T類直接送印、一般送印只能選擇一種列印方式問題
#+ Modifier...: No.190910-00011#1 19/09/11 11084   修正V3.0 GDC直連開啟報表異常問題
#+ Modifier...: No.191127-00011#2 19/11/27 11084   調整日期格式呈現規則：azzi800>aooi100>azzi500
#+ Modifier...: No.200110-00003#1 20/02/07 11084   調整排程無報表資料產生時，寄送通知信主旨規則，若自行設置信件主旨，則以自訂主旨+排程編號寄送，若否則以default主旨。
#+ Modifier...: No.200420-00014#3 20/04/27 11084   配合bpm需求調整簽核規則，新增了簽核關卡名稱選項，3:關卡+姓名/4.關卡+姓名+簽核日期/5.關卡+姓名+簽核日期時間。
#+ Modifier...: No.210329-00008#1 21/09/23 00544   調整報表樣版等待匯出pdf時間
#+ Modifier...: No.230327-00010#1 23/03/27 00544   報表元件使用 g_mail_context 自定信件本文
#+ Modifier...: No.230831-00036#3 23/09/15 00544   使用BPM簽核代理人
#+ Modifier...: No.240130-00011#3 24/03/06 00544   由aoos010參數決定寄信警语
#+ Modifier...: No.240515-00001#1 24/05/15 00544   解決logo取值在不同DB的相容性 
#+ Modifier...: No.230313-00030#2 24/09/27 19425   增加ERP單據狀態卡控
#+ Modifier...: No.240929-00003#1 24/09/29 19425   修复上传logo不显示的问题,取消No.240515-00001#1的调整
#+ Modifier...: No.250224-00011#1 25/02/24 05292   新增支援 v1升級genero 3.20
#+ Modifier...: No.250912-00022#1 25/09/15 14910   新增支援娜娜接口呼叫
#+ Modifier...: No.260327-00001#4 26/03/27 18060   支持T100 AiGP架構運行GR報表
#+ Modifier...: No.260430-00024#1 26/04/30 14910   修正 GR Mail 無法運行問題、報表連續點選列印時會異常、GBC 列印報表會彈 SVG 和 PDF
#end add-point
#add-point:填寫註解說明(客製用) name="main.memo_customerization"

#end add-point
#(ver:6) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:6) --- end ---
 
IMPORT os
#add-point:增加匯入項目 name="main.import"
IMPORT util
IMPORT JAVA com.dsc.t100.report.enc.Pdfenc       #150922-00009#1 add
IMPORT security                                  #170808-00030#1 add
#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
#add-point:增加匯入變數檔 name="global.inc"
GLOBALS "../../cfg/top_report.inc"               #T100報表的環境變數
GLOBALS "../../cfg/top_schedule.inc"             #T100排程的環境變數 
#end add-point
 
{</section>}
 
{<section id="cl_gr.free_style_variable" type="s" >}
#add-point:free_style模組變數(Module Variable) name="free_style.variable"
GLOBALS
   DEFINE ls_outputformat  STRING
   DEFINE li_preview       LIKE type_t.num5       #是否預覽報表
   DEFINE ls_savefilename  STRING                 #歷史報表檔名 
#   DEFINE g_view_cnt       LIKE type_t.num5      #未使用 #170725-00086#1 mark
#   DEFINE g_gr_progress    STRING                #未使用 #170725-00086#1 mark
   DEFINE g_receiver_list  STRING                 #郵件收件人
   DEFINE g_cc_list        STRING                 #郵件副本
   DEFINE g_bcc_list       STRING                 #郵件密件副本
   DEFINE g_adv_list       STRING                 #郵件進階設定 #190107-00075#6 add 
   DEFINE g_sendmail       DYNAMIC ARRAY OF RECORD
             format        STRING,
             subject       STRING,
             context       STRING,
             receiver      STRING,
             cc            STRING,
             bcc           STRING
                           END RECORD
   DEFINE g_mail_cnt       LIKE type_t.num5
   DEFINE ls_savefilename1 STRING                 #歷史報表自訂命名 #待規劃
   DEFINE g_addr_1         STRING                 #公司地址1 for表頭表尾呈現
   DEFINE g_addr_2         STRING                 #公司地址2 for表頭表尾呈現
   DEFINE g_addr_3         STRING                 #公司地址3 for表頭表尾呈現 
   DEFINE g_chk_lang_data  BOOLEAN                #161121-00016#1 add 判斷對內對外名稱資料是否有建置
   
END GLOBALS
#end add-point
 
{</section>}
 
{<section id="cl_gr.global_variable" type="s" >}
#add-point:自定義模組變數(Module Variable) name="global.variable"
#樣版資料
type type_g_gzgd       RECORD
   pick                VARCHAR,                   #顯示勾選
   gzgdl002            LIKE gzgdl_t.gzgdl002,     #樣板說明
   gzgd003             LIKE gzgd_t.gzgd003,       #客製否
   gzgd007             LIKE gzgd_t.gzgd002,       #樣板名稱        
   gzgd002             LIKE gzgd_t.gzgd002,       #樣板代號
   gzgd000             LIKE gzgd_t.gzgd000,       #報表樣板ID
   gzgd005             LIKE gzgd_t.gzgd005,       #用戶
   gzgd004             LIKE gzgd_t.gzgd004,       #角色   
   gzgd008             LIKE gzgd_t.gzgd008,       #預設樣板
   gzgd009             LIKE gzgd_t.gzgd009,       #列印簽核
   gzgd010             LIKE gzgd_t.gzgd010,       #列印簽核位置
   sort_num            LIKE type_t.num5           #排序碼(越大越優先)
   END RECORD
#直接送印設定
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
   gzgd000             LIKE gzgd_t.gzgd000,       #報表樣板ID
   gzgd001             LIKE gzgd_t.gzgd001,       #報表元件代號
   gzgd002             LIKE gzgd_t.gzgd002,       #樣板代號
   gzgd003             LIKE gzgd_t.gzgd003,       #客製
   gzgd004             LIKE gzgd_t.gzgd004,       #角色
   gzgd005             LIKE gzgd_t.gzgd005,       #使用者
   gzgd008             LIKE gzgd_t.gzgd008,       #預設樣板
   gzgd009             LIKE gzgd_t.gzgd009,       #列印簽核
   gzgd010             LIKE gzgd_t.gzgd010,       #簽核位置
   gzgdl002            LIKE gzgdl_t.gzgdl002      #樣板說明
   END RECORD

DEFINE g_gzgd   DYNAMIC ARRAY OF  type_g_gzgd            
DEFINE g_rec_b              LIKE type_t.num10
DEFINE g_sql                STRING
DEFINE g_mail_flag          STRING
DEFINE g_save_flag          STRING
DEFINE g_gr_times           LIKE type_t.num5
DEFINE g_savefilename       STRING
DEFINE g_mail_attach        STRING                 #記錄mail附件檔名
DEFINE g_mail_attach_win    STRING                 #window 
DEFINE g_gr_filename        STRING                 #window
DEFINE g_apr_cnt            LIKE type_t.num5       #紀錄執行列印簽核的次數 
#DEFINE g_bg_cnt             LIKE type_t.num5      #紀錄背景列印簽核的次數  #170104-00023#1 mark
DEFINE g_ac                 LIKE type_t.num10 
DEFINE g_gzgd000            LIKE gzgd_t.gzgd000

DEFINE g_ref_fields         DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
DEFINE g_ref_vars           DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
DEFINE g_rtn_fields         DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
#UI變數
DEFINE gwin_curr            ui.Window                     #Current Window
DEFINE gfrm_curr            ui.Form                       #Current Form
DEFINE gdig_curr            ui.Dialog

#直接送印變數
DEFINE g_dp_flag            STRING                 #是否有設定直接送印
DEFINE g_dp                 type_g_gzgn
DEFINE g_erpid              STRING                 #產品別T100ERP/T100EXT #160608-00012#1 add
DEFINE g_sampledata_path    STRING                 #Sampledata路徑        #160608-00012#1 add
#160912-00001#1-2 add(s)
DEFINE g_4rp_path           STRING
DEFINE g_run_cmd            STRING,
       g_run_result         BOOLEAN,
       g_err_msg            STRING
DEFINE g_errlog      DYNAMIC ARRAY OF STRING       
#160912-00001#1-2 add(e)
DEFINE g_4rp_dir            STRING                 #170406-00030#1 add #4rp路徑
DEFINE g_mail_resp          STRING                 #170725-00086#1 add #下載郵件狀態
DEFINE g_attach_desc        STRING                 #170725-00086#1 add #郵件附件說明/URL
DEFINE g_sampledata_flag    LIKE type_t.chr1       #170808-00030#1 add #是否有選擇sampledata
DEFINE g_sampledata_ac      LIKE type_t.num5       #170808-00030#1 add #產生sampledata時選擇的樣板序號
DEFINE g_vn_flag            LIKE type_t.chr1       #180123-00036#1 add #識別是否選擇中越樣板
DEFINE g_str                STRING                 #180213-00002#1 add
DEFINE g_gzgd002_flag       LIKE type_t.chr1       #180213-00002#1 add
DEFINE gre_private_dir      STRING                 #181213-00050#1 add
DEFINE gre_public_dir       STRING                 #181213-00050#1 add
DEFINE g_template_info      STRING                 #190503-00009#1 add #紀錄樣板相關資訊
DEFINE g_separator_str      STRING                 #190606-00010#1 add #Mail寄送資訊分隔符號
DEFINE g_send_info          STRING                 #190624-00029#1 add #紀錄寄送結果資訊
DEFINE g_scc_flag           LIKE type_t.chr1       #190705-00001#1 add #記錄撈取是否已經POP錯誤訊息
#end add-point
 
{</section>}
 
{<section id="cl_gr.other_dialog" type="s" >}

 
{</section>}
 
{<section id="cl_gr.other_function" readonly="Y" type="s" >}

################################################################################
# Descriptions...: GR列印視窗
# Memo...........:
# Usage..........: LET handler = cl_gr_handler()
#                  
# Input parameter: none
# Return code....: om.SaxDocumentHandler    SAX文件處理器物件   
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_handler()
   DEFINE l_ac              LIKE type_t.num5
   DEFINE l_cnt             LIKE type_t.num5
   DEFINE lc_gzgd007        LIKE gzgd_t.gzgd007     #樣板名稱(4rp)
   DEFINE HANDLER           om.SaxDocumentHandler  
   DEFINE ls_path           STRING
   DEFINE ls_path_mnt       STRING                  #樣板mount目錄(/T100_gr),170406-00030#1改讀取報表主機路徑
   DEFINE ls_path_win       STRING                  #樣板報表主機目錄(D:/T100_gr)
   DEFINE ls_path_erp       STRING                  #樣板ERP主機目錄($MOD/4rp/)  #170406-00030#1 add for check paper size
   DEFINE ls_dir            STRING
   DEFINE l_dialog          ui.Dialog
   DEFINE l_sampledata_flag LIKE type_t.chr1
   DEFINE l_rootnode	       om.DomNode
   DEFINE l_nodes	          om.NodeList
   DEFINE ls_output         STRING                  #輸出格式
   DEFINE l_i               LIKE type_t.num10
   DEFINE l_save            STRING                  #是否留存歷史報表(Y/N)
   DEFINE l_str_gzdel       LIKE gzdel_t.gzdel003   #報表元件代號說明
   DEFINE l_count           INTEGER                 #計算子報表數量
   DEFINE lc_gzgd000        LIKE gzgd_t.gzgd000     #報表id,呼叫LayoutEdit用
   DEFINE l_row             INTEGER                 #紀錄列數
   DEFINE l_action          STRING                  #儲存使用者是否按下Layout edit按鈕
   DEFINE l_svg_name        STRING
   DEFINE l_svg_file        base.Channel            #歷史報表 #create new svgFile
   DEFINE l_delete          SMALLINT                #歷史報表 刪除檔案返回之參數
   DEFINE l_ooef013         LIKE ooef_t.ooef013     #日期格式代碼
   DEFINE l_ooef023         LIKE ooef_t.ooef023     #數值格式代碼

   DEFINE l_date_fmt        STRING                  #日期格式
   DEFINE l_num_fmt_all     STRING                  #數值格式
   DEFINE l_drive_str       STRING                  #存fgl_getenv("TEMPLATEDRIVE")的字串
   DEFINE l_l ,i            LIKE type_t.num5
   DEFINE l_chk             LIKE type_t.chr1        #判斷tiptop群組
   DEFINE l_chmodname       STRING            
   DEFINE l_grdatetime      DATETIME YEAR  TO SECOND   
   DEFINE l_sysdatetime     DATETIME YEAR  TO SECOND 
   DEFINE l_timediff        INTERVAL SECOND TO FRACTION(5)
   DEFINE ls_4st_path       STRING                  #150526-00012#2 add
   DEFINE mc_lang           STRING                  #150526-00012#2 add
   DEFINE l_gzgo004         LIKE gzgo_t.gzgo004     #150803-00008#3 add
   DEFINE l_rs_prt_flag     LIKE type_t.chr1        #Server直接送印印表機名稱 #150921-00060#1 add
   DEFINE l_gzcbl004        LIKE gzcbl_t.gzcbl004   #日期格式 #160817-00023#1 add
   DEFINE l_bgprinter       LIKE gzpe_t.gzpe005     #背景排程Server送印印表機名稱 #170710-00060#1 add
   DEFINE l_result          BOOLEAN                 #170808-00030#1 add
   DEFINE l_gre_ver         STRING                  #GRE版本 #171211-00055#1 add
   DEFINE l_gzpb009         LIKE gzpb_t.gzpb009     #180719-00037#1 add #紀錄樣板
   #181213-00050#1 add(s)
   DEFINE l_uuid            STRING
   DEFINE l_grehttpd_dir    STRING                  #GREHttpd根目錄
   DEFINE l_separator       STRING                  #紀錄os.Path.separator分隔符號
   #181213-00050#1 add(e)   
   DEFINE l_gzxa020   LIKE  gzxa_t.gzxa020          #190117-00039#1  add 用戶保留列印視窗設定值
   DEFINE l_site_t          STRING                  #190327-00044#1 add   #備份g_site用
   #190503-00009 add(s)
   DEFINE l_gzgd002         LIKE gzgd_t.gzgd002     #紀錄樣板資訊
   DEFINE l_gzgd004         LIKE gzgd_t.gzgd004     #紀錄樣板資訊
   DEFINE l_gzgd005         LIKE gzgd_t.gzgd005     #紀錄樣板資訊
   #190503-00009 add(e)
   DEFINE l_frontend        STRING                  #190910-00011#1 add
   #260327-00001#4 add(s)
   DEFINE lc_str        STRING       
   DEFINE la_param   RECORD
      prog              STRING,
      param             DYNAMIC ARRAY OF STRING
                     END RECORD 
   DEFINE ls_filename   STRING
   DEFINE l_url         STRING
   DEFINE l_ostype        STRING
   #260327-00001#4 add(e)
   WHENEVER ERROR CONTINUE   

   
   #151027-00003#1 add -(s)
   IF cl_rpt_chk_env_exist() THEN
#      LET INT_FLAG = TRUE    #151201-00015#1 mark
      LET g_rep_flag = TRUE   #151201-00015#1 add
      LET g_err_msg = "Please Check Env Exist! "   #160912-00001#1-2 add
      CALL cl_gr_logfile()                         #160912-00001#1-2 add
      RETURN NULL
   END IF   
   #151027-00003#1 add -(e)

   LET g_erpid = NVL(FGL_GETENV("ERPID"),"T100ERP")   #160608-00012#1 add
   LET l_frontend = ui.Interface.getFrontEndName()    #190910-00011#1 add
   #190327-00044#1 add(s)
   #BPM流程串接時g_site可能會是ALL，改用使用者預設site
   LET l_site_t = ""
   IF g_site = "ALL" THEN
      LET l_site_t = g_site
      SELECT gzxa007 INTO g_site FROM gzxa_t
       WHERE gzxaent = g_enterprise AND gzxa001 = g_account
   END IF
   #190327-00044#1 add(e)
        
   #170808-00030#1 add(s)
   #檢查是否有產生sampledata，若有則將相關檔案一併壓縮，提供下載
   IF g_sampledata_flag = "Y" THEN
      LET g_sampledata_flag = "N"
      CALL cl_gr_sampledata_zip() RETURNING l_result
      #190321-00031#1 add(s)
      IF cl_gr_get_user_gzxa020() = "N" THEN
         LET g_rep_flag = TRUE
         RETURN NULL
      END IF
      #190321-00031#1 add(e)
   END IF
   #170808-00030#1 add(e)

   #260430-00024#1 add(s) 變數初始化
   IF (ls_outputformat <> "9" OR g_bgjob <> "Y") THEN
      LET ls_outputformat = NULL
   END IF
   #260430-00024#1 add(e)

   #取得直接送印參數
   LET g_dp_flag = 'N'
#   IF g_action_choice = "quickprint" THEN                            #150921-00060#1 add   #點擊"快速列印"時才檢查直接送印  #160817-00023#1 MARK
#   IF g_action_choice = "quickprint"  OR g_bgjob="T" THEN             #160817-00023#1 add 增加T作業判斷   #190710-00007#1 mark
#   IF g_action_choice = "quickprint"  OR (g_bgjob="T" AND g_actdefault = "quickprint") THEN                                 #250912-00022#1 mark
   IF g_action_choice = "quickprint"  OR (g_bgjob="T" AND g_actdefault = "quickprint") OR FGL_GETENV("ATHENA") == "1" THEN   #250912-00022#1 add
      CALL cl_gr_direct_print()
   END IF                                             #150921-00060#1 add 

   IF g_sendmail.getLength() > 0 AND g_mail_cnt = g_sendmail.getLength() THEN
      CALL cl_gr_javamail_prepare(g_mail_cnt)
      LET g_mail_flag = FALSE
      CALL g_sendmail.clear()
      LET g_mail_cnt = 0
#      LET g_rep_flag = TRUE #190117-00039#1 add 寄完郵件離開視窗   #190321-00031#1 mark
      #190321-00031#1 add(s)
      IF g_save_flag = "N" AND cl_gr_get_user_gzxa020() ="N" THEN
         #寄完郵件離開視窗
         LET g_rep_flag = TRUE   
      END IF
      #190321-00031#1 add(e)
      #190624-00029#1 add(s)
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code =  "lib-00648"
      LET g_errparam.replace[1] = "\n\n",g_send_info 
      LET g_errparam.popup = TRUE
      CALL cl_err()
      LET g_send_info = NULL
      #190624-00029#1 add(e)
   END IF

   #歷史報表檔名檢查
   IF g_save_flag = "S" THEN
      LET g_save_flag = "N"
      #LET l_chmodname=os.Path.join(os.Path.join(FGL_GETENV("MNT4RP"),"history"),os.Path.basename(g_savefilename))              ##150916-00002#1 mark   
#      LET l_chmodname=os.Path.join(os.Path.join(cl_rpt_get_env_global("MNT4RP"),"history"),os.Path.basename(g_savefilename))    ##150916-00002#1 add   #170725-00086#1 mark
      LET l_chmodname=os.Path.join(os.Path.join(cl_rpt_get_env_global("PATH4RP"),"history"),os.Path.basename(g_savefilename))    ##150916-00002#1 add   #170725-00086#1 add
      LET l_sysdatetime = CURRENT  YEAR  TO SECOND  #系統現在時間
      CALL os.Path.mtime(l_chmodname) RETURNING l_grdatetime  #pdf最後修改時間         
      LET l_timediff = l_sysdatetime - l_grdatetime    
      DISPLAY 'l_chmodname=',l_chmodname
      #直接送印時，儲存完歷史報表後須將g_save_flag清空，避免執行快速送印時，設置多筆樣板沒彈出選擇視窗
      IF g_action_choice = 'quickprint' THEN LET g_save_flag = NULL END IF   #190710-00007#1 add
      #190321-00031#1 add(s)
#      IF cl_gr_get_user_gzxa020() = "N" THEN   #190710-00007#1 mark
      #當直接送印時，不保留列印詢問視窗
      IF cl_gr_get_user_gzxa020() = "N" OR g_dp_flag = "Y" THEN   #190710-00007#1 add
         LET g_rep_flag = TRUE
         RETURN NULL
      END IF
      #190321-00031#1 add(e)
      #170725-00086#1 mark(s)   #因透過URL存取,無法判斷報表主機上的檔案是否已存在
#      IF os.Path.exists(l_chmodname) AND os.Path.size(l_chmodname)>0 AND l_timediff>216000 THEN #歷史報表檔案存在，且size>0，且上次修改時間超過6小時
#          LET g_err_msg = "LINE 261 History Report is exist."   #160912-00001#1-2 add
#          CALL cl_gr_logfile()                         #160912-00001#1-2 add
#          RETURN NULL 
#      END IF 
      #170725-00086#1 mark(e)
   END IF  

   #170710-00060#1 add(s)

   IF g_bgjob = "Y" THEN
      SELECT gzpe005 INTO l_bgprinter FROM gzpe_t WHERE gzpeent = g_enterprise AND gzpe001 = g_schedule_id AND gzpe002 = g_schedule_seq
         AND gzpe003 = 'Y'
      IF NOT cl_null(l_bgprinter) THEN
         LET g_rs_printer = l_bgprinter
      END IF
   END IF
   #170710-00060#1 add(e)

   #150921-00060#1 add(s)
   #當g_rs_printer有值時,直接視為Server直接送印
   #IF NOT cl_null(g_rs_printer) AND g_action_choice = "quickprint" THEN   #170710-00060#1 mark
   IF NOT cl_null(g_rs_printer) AND g_action_choice = "quickprint" OR NOT cl_null(l_bgprinter) THEN   #170710-00060#1 add
      LET g_bgjob = "Y"
      IF g_schedule_d.getLength() > 0 THEN
         LET l_i = g_schedule_d.getLength() +1
      ELSE
         LET l_i = 1
      END IF
         #170710-00060#1 add(s)

      LET g_schedule_d[l_i].gzpe003 = "Y"
      LET g_schedule_d[l_i].gzpe008 = g_rs_printer CLIPPED
      LET g_schedule_d[l_i].gzpe009 = 1
   END IF   
   LET l_rs_prt_flag = "N"
   #150921-00060#1 add(e)
   
#   IF g_mail_flag OR g_save_flag="Y"   OR (g_bgjob = "Y" AND FGL_GETENV("FGLGUI")='0')THEN  #背景作業   
   IF g_mail_flag OR g_save_flag="Y" OR g_bgjob = "Y" THEN   #160817-00023#1 mark   #190710-00007#1 remark
#   IF g_mail_flag OR g_save_flag="Y" OR g_bgjob = "Y" OR (g_bgjob ="T" AND g_dp_flag="Y") THEN    #160817-00023#1 add T類直接送印   #190710-00007#1 mark
      #背景作業
      IF g_bgjob = "Y" THEN                                                                 #160817-00023#1 mark   #190710-00007#1 remark
#      IF g_bgjob = "Y" OR (g_bgjob ="T" AND g_dp_flag="Y") THEN                              #160817-00023#1 add   #190710-00007#1 mark 
         LET g_bg_cnt = g_bg_cnt + 1                                                         
         #IF g_bg_cnt <= 1 THEN                                                              #160817-00023#1 mark 
         #IF g_bg_cnt <= 1 AND g_bgjob <> "T" THEN                                            #160817-00023#1 add  #161103-00014#1 mark
         IF g_bg_cnt <= 1 THEN                                                                #161103-00014#1 add
            LET lc_gzgd007 = NULL
            LET l_ac = 0
            CALL cl_gr_b_fill()


            IF g_rep_flag = TRUE THEN RETURN NULL END IF   #180123-00036#1 add
            #180719-00037#1 add(s)
            SELECT gzpb009 INTO l_gzpb009 
              FROM gzpb_t 
             WHERE gzpbent = g_enterprise AND gzpb001 = g_schedule_id AND gzpb002 = g_schedule_seq
            IF cl_null(g_reportid) THEN
               LET g_reportid = l_gzpb009
            END IF       
            #180719-00037#1 add(e)
            #有設定樣版名稱時要比對樣板名稱 
           
            IF NOT cl_null(g_reportid) THEN #排程
               FOR l_i = 1 TO g_gzgd.getLength() 
                  IF g_gzgd[l_i].gzgd000 = g_reportid THEN #排程
                     LET l_ac = l_i
                     LET lc_gzgd007 = g_gzgd[l_i].gzgd007
                     EXIT FOR
                  END IF
               END FOR
               #190503-00009#1 add(s)
               SELECT gzgd002,gzgd004,gzgd005 INTO l_gzgd002,l_gzgd004,l_gzgd005 FROM gzgd_t WHERE gzgd000 = g_reportid
               LET g_template_info = "Template gzgd000:", g_reportid," ,gzgd002:",l_gzgd002," ,gzgd004:",l_gzgd004," ,gzgd005:",l_gzgd005
               #190503-00009#1 add(e)
            ELSE #未設定樣板名稱時抓第一筆
               IF g_gzgd.getLength() >= 1 THEN
                  LET l_ac = 1
                  LET lc_gzgd007 = g_gzgd[l_ac].gzgd007
               END IF
               LET g_template_info = "Template gzgd000:",g_gzgd[l_ac].gzgd000," ,gzgd002:",g_gzgd[l_ac].gzgd002," ,gzgd004:",g_gzgd[l_ac].gzgd004," ,gzgd005:",g_gzgd[l_ac].gzgd005   #190503-00009#1 add
            END IF
            DISPLAY g_template_info        #190503-00009#1 add
            IF g_bgjob <> "T" THEN         #161103-00014#1 add
               IF l_ac > 0 THEN
                  LET li_preview = 0
                  IF NOT cl_null(l_bgprinter) THEN CALL g_schedule_d.Clear() END IF    #170710-00060#1 add
                  CALL cl_gr_mail()
               ELSE
   #               LET INT_FLAG = TRUE    #151201-00015#1 mark
                  LET g_rep_flag = TRUE   #151201-00015#1 add
                  LET g_bg_cnt = 0
                  LET g_err_msg = "LINE 320 There is no report template."   #160912-00001#1-2 add
                  CALL cl_gr_logfile()                             #160912-00001#1-2 add
                  RETURN NULL
               END IF
            END IF                          #161103-00014#1 add
         END IF      
         IF g_bg_cnt = 1 AND g_sendmail.getLength() = 0 THEN
            LET g_save_flag = "Y"
         END IF

         #計算背景作業執行的次數,次數滿之後跳出cl_gr_handler()
         IF g_bg_cnt > g_sendmail.getLength() AND g_sendmail.getLength() >= 0 AND g_bg_cnt > 1 THEN
#            LET INT_FLAG = TRUE    #151201-00015#1 mark
            LET g_rep_flag = TRUE   #151201-00015#1 add
            LET g_bg_cnt = 0
#            RETURN NULL   #150921-00060#1 mark
            #150921-00060#1 add(s)
            FOR l_i = 1 TO g_schedule_d.getLength()
               IF g_schedule_d[l_i].gzpe003 = "Y" THEN
                  LET ls_outputformat = "S"
                  LET g_dp.gzgn006 = g_schedule_d[l_i].gzpe008
                  LET g_dp.gzgn007 = g_schedule_d[l_i].gzpe009
                  LET l_rs_prt_flag = "Y"
#                  LET INT_FLAG = FALSE    #151201-00015#1 mark
                  LET g_rep_flag = FALSE   #151201-00015#1 add
                  EXIT FOR
               END IF
            END FOR
            
            IF l_rs_prt_flag <> "Y" THEN
               LET g_err_msg = "LINE 349 l_rs_prt_flag <> 'Y'"   #160912-00001#1-2 add
               CALL cl_gr_logfile()                              #160912-00001#1-2 add            
               RETURN NULL
            END IF   
            #150921-00060#1 add(e)            
         END IF
      END IF
      IF g_save_flag="Y" THEN
         LET li_preview = 0
         LET l_sampledata_flag = "N"
         LET l_ac = 1
         LET g_dp_flag = "N"   #190710-00007#1 add
      END IF

      IF g_mail_flag THEN
         IF g_mail_cnt <= g_sendmail.getLength() THEN
            IF g_mail_cnt > 0 THEN              
               CALL cl_gr_javamail_prepare(g_mail_cnt)
            END IF
            LET g_mail_cnt = g_mail_cnt + 1

            # 迴圈指定handler, 跑完以後再回到cl_gr_handler送信
            CASE g_sendmail[g_mail_cnt].format
               WHEN "1" LET ls_outputformat = "2"  #PDF 
               WHEN "2" LET ls_outputformat = "3"  #XLS
               WHEN "3" LET ls_outputformat = "4"  #XLS 每頁獨立工作簿
               WHEN "4" LET ls_outputformat = "5"  #XLSX
               WHEN "5" LET ls_outputformat = "6"  #XLSX 每頁獨立工作簿
               WHEN "6" LET ls_outputformat = "7"  #RTF
               WHEN "7" LET ls_outputformat = "8"  #HTML
               OTHERWISE LET ls_outputformat = "2" #格式預設為PDF
            END CASE
            LET li_preview = 0
            LET l_sampledata_flag = "N"
            #LET ls_savefilename = os.Path.join(FGL_GETENV("MNT4RP"),"tmp")             #150916-00002#1 mark
#            LET ls_savefilename = os.Path.join(cl_rpt_get_env_global("MNT4RP"),"tmp")   #150916-00002#1 add   #170725-00086#1 mark
            LET ls_savefilename = os.Path.join(cl_rpt_get_env_global("PATH4RP"),"tmp")   #150916-00002#1 add   #170725-00086#1 add
            LET l_ac = 1
         #150921-00060#1 add(s)
         ELSE           
            FOR l_i = 1 TO g_schedule_d.getLength()
               IF g_schedule_d[l_i].gzpe003 = "Y" THEN
                  LET ls_outputformat = "S"
                  LET g_dp.gzgn006 = g_schedule_d[l_i].gzpe008
                  LET g_dp.gzgn007 = g_schedule_d[l_i].gzpe009
                  LET l_rs_prt_flag = "Y"
                  LET g_save_flag = "N"
                  LET l_sampledata_flag = "N"                  
               END IF
            END FOR      
         #150921-00060#1 add(e)             
         END IF
      END IF
      IF g_ac > 0 THEN        
        LET l_ac = g_ac
      END IF      
   ELSE
      #非直接送印要開GR列印視窗
      IF g_dp_flag = 'N' THEN
         #150526-00012#2 add(s)
         #T類作業需調整開窗Style
         IF g_bgjob = 'T' THEN
            IF cl_null(mc_lang) THEN
               LET mc_lang = g_lang CLIPPED
            END IF            
            LET ls_4st_path = os.Path.join(os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("ERP"),"cfg"),"4st"),mc_lang),"style_r_tpl-H1_blue.4st")
            CALL ui.Interface.loadStyles(ls_4st_path)
         END IF
         #150526-00012#2 add(e)
         
         LET l_rootnode = ui.Interface.getRootNode()
         LET l_nodes = l_rootnode.selectByPath("//Window[@name=\"cl_gr_w\"]")
         IF l_nodes.getLength() <= 0 THEN
            OPEN WINDOW cl_gr_w WITH FORM cl_ap_formpath("lib","cl_gr_s01")         
               ATTRIBUTES(STYLE="layoutwin")
         END IF

         CALL cl_ui_init()

         LET ls_outputformat = "1"
         LET li_preview = 1
         LET l_sampledata_flag = "N"
#         LET ls_savefilename = os.Path.join(FGL_GETENV("TEMPDIR"),"output")   #170725-00086#1 mark
         LET l_count = 1
         
         #150803-00008#3 add(s)
         #允許列印方式權限控管
#         CALL cl_gr_get_gzgo004("","","") RETURNING l_gzgo004    #150901-00021#1 mark
#         CALL cl_gr_format_combo(l_gzgo004,'outputformat','G')   #150901-00021#1 mark     
         CALL cl_rpt_get_gzgo004("","","") RETURNING l_gzgo004    #150901-00021#1 add
         CALL cl_rpt_format_combo(l_gzgo004,'outputformat','G')   #150901-00021#1 add    
         #150803-00008#3 add(e)
       
         CALL cl_gr_chk_tiptopgroup() RETURNING l_chk  #判斷tiptop群組
         IF (g_user = 'tiptop') OR l_chk = 'Y' THEN   
            #CALL cl_set_comp_visible("layout_edit",TRUE) #140616 mark
            CALL cl_set_comp_visible("layout_edit",FALSE) #140616 先寫死
#            CALL cl_set_comp_visible("btn_sampledata",TRUE)   #170808-00030#1 add   #170815-00008#1 mark
#            CALL cl_set_comp_visible("sampledata",TRUE)      #170808-00030#1 mark
         ELSE 
            CALL cl_set_comp_visible("layout_edit",FALSE) 
#            CALL cl_set_comp_visible("btn_sampledata",FALSE)   #170808-00030#1 add   #170815-00008#1 mark
#            CALL cl_set_comp_visible("sampledata",FALSE)      #170808-00030#1 mark
         END IF

         DISPLAY ls_outputformat
            TO outputformat
      END IF
      
      LET l_gzxa020 = cl_gr_get_user_gzxa020()   #190117-00039#1  add

      
      #180213-00002#1 mark(s)
#      #檢查g_template是否存在gzgd_t中
#      IF NOT cl_null(g_template) THEN       
##         SELECT COUNT(*) INTO l_cnt FROM gzgd_t WHERE gzgdstus = "Y" AND gzgd001 = g_rep_code AND gzgd002 = g_template   #170525-00105#1 mark
#         SELECT COUNT(1) INTO l_cnt FROM gzgd_t WHERE gzgdstus = "Y" AND gzgd001 = g_rep_code AND gzgd002 = g_template    #170525-00105#1 add
#      ELSE
##         SELECT COUNT(*) INTO l_cnt FROM gzgd_t WHERE gzgdstus = "Y" AND gzgd001 = g_rep_code                            #170525-00105#1 mark
#         SELECT COUNT(1) INTO l_cnt FROM gzgd_t WHERE gzgdstus = "Y" AND gzgd001 = g_rep_code                             #170525-00105#1 add
#      END IF      
      #180213-00002#1 mark(e)

      #180213-00002#1 add(s)
      #檢查g_template是否存在gzgd_t中
      LET g_sql = " SELECT COUNT(1) FROM gzgd_t WHERE gzgdstus = 'Y' AND gzgd001 = '",g_rep_code,"' "
      IF NOT cl_null(g_template) THEN
         LET g_sql = g_sql, " AND gzgd002 = '",g_template,"' "    
      END IF
      
      IF NOT cl_null(g_bg4rp) THEN
         LET g_gzgd002_flag = "Y" 
         #檢查g_bg4rp的內容
         FOR i = 1 TO g_bg4rp.getLength()
            LET l_cnt = g_bg4rp.getIndexOf("gzgd",i)
            IF l_cnt > 0 THEN
               LET g_str = g_bg4rp.subString(l_cnt,l_cnt+6)
               DISPLAY "g_str:",g_str
               IF g_str.trim() <> "gzgd002" THEN
                  LET g_gzgd002_flag = "N"   #sql包含不是gzgd002的內容不使用
               END IF
               LET i = l_cnt
            END IF
         END FOR
         IF g_gzgd002_flag = "Y" THEN
            LET g_sql = g_sql," AND ",g_bg4rp
         END IF      
      END IF
      
      LET l_cnt = 0
      PREPARE cl_gr_gzgd002_pre FROM g_sql
      EXECUTE cl_gr_gzgd002_pre INTO l_cnt      
      #180213-00002#1 add(e)
      
      LET l_gzxa020 = cl_gr_get_user_gzxa020()   #190117-00039#1  add #取得用戶是否保留列印視窗值
      
      
      IF l_cnt <= 0 THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'lib-00111'
         LET g_errparam.extend = g_template
         LET g_errparam.popup = TRUE
         CALL cl_err()

#         LET INT_FLAG = TRUE    #151201-00015#1 mark
         LET g_rep_flag = TRUE   #151201-00015#1 add
         LET g_err_msg = "LINE 470 There is no template."   #160912-00001#1-2 add
         CALL cl_gr_logfile()                               #160912-00001#1-2 add           
         RETURN NULL
      ELSE
         CALL cl_gr_b_fill()
         IF g_rep_flag = TRUE THEN RETURN NULL END IF   #180123-00036#1 add         
      END IF
      
      #非直接送印要開GR列印視窗
      IF g_dp_flag = 'N' THEN
         #190327-00044#1 add(s)
         &ifdef v1
#            IF FGL_GETENV("GUIMODE") = "GWC" THEN   #190910-00011 mark
            IF l_frontend <> "GDC" THEN   #190910-00011 add
               CALL cl_set_comp_entry("outputformat",FALSE)
               LET ls_outputformat = "2"
            END IF
         &endif
         #190327-00044#1 add(e)
      #GWC時期固定使用PDF，但GBC已提供HTML5的報表，故開放選項 #181213-00050#1 mark -(s)
#         IF FGL_GETENV("GUIMODE") = "GWC" THEN            
#            CALL cl_set_comp_entry("outputformat",FALSE)
#            LET ls_outputformat = "2"           
#         END IF
#      
#         #檢查cl_gr_b_fill取出的g_gzgd陣列是否只有一筆, 單筆狀況不顯示選擇樣板畫面 #for GWC
#         IF FGL_GETENV("GUIMODE") = "GWC" AND g_gzgd.getLength() = 1 THEN 
#            LET l_ac = 1  
#         ELSE    #非GWC或GWC但只有一筆樣板
       #GWC時期固定使用PDF，但GBC已提供HTML5的報表，故開放選項 #181213-00050#1 mark -(e)  
            DIALOG ATTRIBUTE(UNBUFFERED)
   
               DISPLAY ARRAY g_gzgd TO s_gzgd.* ATTRIBUTE(COUNT = g_rec_b)
              
                  BEFORE DISPLAY
                     LET l_action = ''
                     LET l_count = ''
                     LET g_gzgd[1].pick = 'Y'                 #初始狀態第一行必勾選
                     IF g_gzgd.getLength() != 1 THEN          #若不只一筆樣板檔
                        FOR l_count = 2 TO g_gzgd.getLength() #其餘筆數均預設不勾選
                           LET g_gzgd[l_count].pick = 'N'
                        END FOR
                           CALL g_gzgd.deleteElement(l_count) #刪除系統自動產生的最後一筆空行
                           LET l_ac = 1
                     #190503-00009#1 add(s)
                     #預先先指定l_ac的值，避免當只有一筆樣板時，DISPLAY樣板資訊時錯誤
                     ELSE
                        LET l_ac = 1
                     #190503-00009#1 add(e)
                     END IF
                     
                  BEFORE ROW
                     IF g_gzgd.getLength() != 1 AND l_ac != 0 THEN       #若不只一筆樣板檔
                        IF l_ac != l_dialog.getCurrentRow("s_gzgd") THEN #選擇新筆資料時
                           LET g_gzgd[l_ac].pick = 'N'                   #將上筆取消勾選
                           LET l_ac = l_dialog.getCurrentRow("s_gzgd")   #取得新筆行數
                           LET g_gzgd[l_ac].pick = 'Y'                   #新筆設置為勾選
                        END IF
                     END IF
    
                  AFTER ROW
                     LET l_ac = l_dialog.getCurrentRow("s_gzgd")               
              END DISPLAY
   
              INPUT ls_outputformat,l_sampledata_flag,l_gzxa020   ##190117-00039#1 add l_gzgx020
                 FROM outputformat,sampledata,gzxa020             ##190117-00039#1 add gzgx020
                 ATTRIBUTE(WITHOUT DEFAULTS)
                 ON CHANGE outputformat
                    IF ls_outputformat = "Mail" THEN 
                       LET li_preview = "0"
                    END IF   
                 ##190117-00039#1 add -(s)    
                 #當gzxa020改變，要回寫原欄位 
                 ON CHANGE gzxa020
                    
                       UPDATE gzxa_t
                          SET gzxa020 = l_gzxa020
                        WHERE gzxaent = g_enterprise and gzxa003 =g_user
                   
                 ##190117-00039#1 add -(e)                    

              END INPUT
   
              BEFORE DIALOG
                 LET l_dialog = ui.Dialog.getCurrent()
   
              ON IDLE g_idle_seconds
                 CALL cl_on_idle()
    
#              ON ACTION set_default   #170815-00008#1 mark
           
#              ON ACTION layout_edit #新增按鈕串到Layout Editor
#                 LET l_ac = l_dialog.getCurrentRow("s_gzgd")
#                 LET l_action = 'layout_edit'
#                 ACCEPT DIALOG

              ON ACTION btn_accept
                 #190503-00009#1 add(s)
                 LET g_template_info = "Template gzgd000:",g_gzgd[l_ac].gzgd000," ,gzgd002:",g_gzgd[l_ac].gzgd002," ,gzgd004:",g_gzgd[l_ac].gzgd004," ,gzgd005:",g_gzgd[l_ac].gzgd005
                 DISPLAY g_template_info
                 #190503-00009#1 add(e)
                 LET g_scc_flag = "N"   #初始scc錯誤訊息POP flag   #190705-00001#1 add
                  ACCEPT DIALOG
   
              ON ACTION btn_cancel
#                  LET INT_FLAG = TRUE    #151201-00015#1 mark
                  LET g_rep_flag = TRUE   #151201-00015#1 add
                  EXIT DIALOG
    
              #170815-00008#1 mark(s)
#              #170808-00030#1 add(s)
#              ON ACTION btn_sampledata
#                 LET l_sampledata_flag = "Y"
#                 LET ls_outputformat = "1"
#                 ACCEPT DIALOG
#              #170808-00030#1 add(e)
              #170815-00008#1 mark(e)
   
              ON ACTION CLOSE
#                  LET INT_FLAG = TRUE    #151201-00015#1 mark
                  LET g_rep_flag = TRUE   #151201-00015#1 add
                  EXIT DIALOG
   
#              ON ACTION set_default
#                  IF cl_gre_set_default(l_ac) THEN
#                     LET l_ac = 1
#                     CALL DIALOG.setCurrentRow("s_gzgd", l_ac)
#                  END IF
            END DIALOG
        # END IF #GWC   ##181213-00050#1 mark 
      END IF    #g_dp_flag = 'N'
   END IF
   
#   IF INT_FLAG THEN    #151201-00015#1 mark
   IF g_rep_flag THEN   #151201-00015#1 add
      
      LET handler = NULL
      #判斷是郵件或排程時可以跑2次
      #190117-00039#1  add -(s)
      DISPLAY "ls_outputformat:",ls_outputformat
      DISPLAY "g_rep_flag:",g_rep_flag
      IF  (ls_outputformat <> "9" OR ls_outputformat <> "") THEN          #郵件或排程不走這段
         IF cl_gr_get_user_gzxa020() ="N" THEN  #當沒有設定個人的列印視窗是保留時，預設是使用完關窗          
             CALL cl_gr_history_save() RETURNING l_save   #判斷是否需留存歷史報表
             IF l_save <> "Y" THEN   #不用做歷史報表
                LET  g_rep_flag = TRUE
                LET g_save_flag = NULL   #清掉歷史報表變數
                LET g_gr_times = 0
             END IF
          END IF
      
       END IF 
      #190117-00039#1  add -(e)       
   ELSE
   
      IF g_dp_flag = 'N' THEN
         FOR i= 1 to g_gzgd.getLength()
            IF g_gzgd[i].pick = 'Y' THEN
               LET l_ac = i
               EXIT FOR
            END IF
         END FOR
         #190117-00039#1  add -(s)
         IF ((g_mail_flag = 0 OR cl_null(g_mail_flag)) AND (ls_outputformat <> "9" OR ls_outputformat <> "") )   #判斷非郵件，走第二回要跳出
          OR (g_mail_flag ="1" AND g_mail_cnt = 0)   THEN   #判斷郵件第二回時，還不能跳出 

            IF cl_gr_get_user_gzxa020() ="N" THEN  #當沒有設定個人的列印視窗是保留時，預設是使用完關窗
               LET  g_rep_flag = TRUE
               #190321-00031 mark(s)               
#               LET g_save_flag = NULL   #清掉歷史報表變數
#               LET g_gr_times = 0
               #190321-00031 mark(e)
             END IF
         END IF
         #190117-00039#1  add -(e)          
      ELSE
         #若為直接送印，就將直接送印的資料填入陣列中
         LET l_ac = 1
         LET g_gzgd[l_ac].pick = 'Y'               #顯示勾選
         LET g_gzgd[l_ac].gzgd000 = g_dp.gzgd000   #報表樣板ID
         LET g_gzgd[l_ac].gzgd002 = g_dp.gzgd002   #樣板代號
         LET g_gzgd[l_ac].gzgd003 = g_dp.gzgd003   #客製
         LET g_gzgd[l_ac].gzgd004 = g_dp.gzgd004   #角色
         LET g_gzgd[l_ac].gzgd005 = g_dp.gzgd005   #使用者
         LET g_gzgd[l_ac].gzgd007 = g_dp.gzgn004   #樣板名稱(4rp)
         LET g_gzgd[l_ac].gzgd009 = g_dp.gzgd009   #列印簽核
         LET g_gzgd[l_ac].gzgd010 = g_dp.gzgd010   #簽核位置
         LET g_gzgd[l_ac].gzgdl002 = g_dp.gzgdl002 #樣板說明
         LET g_gzgd[l_ac].sort_num = ''            #排序碼(越大越優先)
         LET li_preview = 1   #150904-00005#4 add
         LET lc_gzgd007 = g_gzgd[l_ac].gzgd007     #181123-00035#1 add
      END IF
      
      IF cl_null(lc_gzgd007) THEN
         LET lc_gzgd007 = g_gzgd[l_ac].gzgd007 CLIPPED
      END IF
      
      #180123-00036#1 add(s)
      LET g_vn_flag = "N"      
      SELECT gzgd012 INTO g_vn_flag FROM gzgd_t WHERE gzgd000 = g_gzgd[l_ac].gzgd000 AND gzgdstus = 'Y'
      #180123-00036#1 add(e)
      
      #當g_gzgd筆數為1時, 不會進入畫面, 故人工指定到第1筆
      IF g_gzgd.getLength() = 1 THEN
         LET l_ac = 1
      END IF
     
      LET g_ac = l_ac 
     
      LET g_gzgd000 = g_gzgd[l_ac].gzgd000
      IF g_apr_cnt = 0 THEN
         CALL cl_gr_get_apr_loc(g_gzgd[l_ac].gzgd009,g_gzgd[l_ac].gzgd010) RETURNING g_grPageHeader.apr_loc   #簽核列印位置
      END IF
           
      CALL cl_gr_get_apr_loc(g_gzgd[l_ac].gzgd009, g_gzgd[l_ac].gzgd010) RETURNING g_grPageHeader.apr_loc   #簽核列印位置
     
      #明細類報表組簽核字串      
      LET g_apr_cnt = g_apr_cnt + 1
     
      #若有設定直接送印，則樣板路徑抓直接送印設定的變數
      #170406-00030#1 mark(s)
##      LET ls_path_mnt = cl_gr_get_4rpdir(g_rep_code,g_gzgd[l_ac].gzgd003,g_rlang,"M")      #150901-00021#1 mark
#      IF g_erpid = "T100ERP" THEN                                                           #160608-00012#1 add
#         LET ls_path_mnt = cl_rpt_get_4rpdir(g_rep_code,g_gzgd[l_ac].gzgd003,g_rlang,"M")   #150901-00021#1 add
#      ELSE                                                                                  #160608-00012#1 add
#         LET ls_path_mnt = cl_rpt_get_4rpdir(g_rep_code,g_gzgd[l_ac].gzgd003,g_rlang,"L")   #160608-00012#1 add
#      END IF                                                                                #160608-00012#1 add
      #170406-00030#1 mark(e)
      
      LET ls_path_erp = cl_rpt_get_4rpdir(g_rep_code,g_gzgd[l_ac].gzgd003,g_rlang,"L")       #170406-00030#1 add
      
#      IF NOT os.Path.exists(ls_path_mnt) THEN                                               #170406-00030#1 mark
      IF NOT os.Path.exists(ls_path_erp) THEN                                                #170406-00030#1 add
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'azz-00676'
#         LET g_errparam.extend = ls_path_mnt                                                #170406-00030#1 mark
         LET g_errparam.extend = ls_path_erp                                                 #170406-00030#1 add
         LET g_errparam.popup = TRUE
         CALL cl_err()
     
#         DISPLAY "ls_path_mnt:",ls_path_mnt                                                 #170406-00030#1 mark
         DISPLAY "ls_path_erp:",ls_path_erp                                                  #170406-00030#1 add
#         EXIT PROGRAM -1   #150904-00005#4 mark
         CALL cl_log_ap_exiterror("-1","azz-00676")   #150904-00005#4 add
      END IF  
      #170406-00030#1 mark(s)
#      LET ls_dir  = ls_path_mnt
#      LET ls_path_mnt = os.Path.join(ls_path_mnt,lc_gzgd007||".4rp")
#      LET g_4rp_path = ls_path_mnt   #160912-00001#1-2 add      
#      DISPLAY 'ls_path_mnt: ', ls_path_mnt
      #170406-00030#1 mark(s)
      #170406-00030#1 add(s)
      LET ls_dir  = ls_path_erp
      LET ls_path_erp = os.Path.join(ls_path_erp,lc_gzgd007||".4rp")
      LET g_4rp_path = ls_path_erp   #160912-00001#1-2 add      
      #170406-00030#1 add(s)
      
#      IF os.Path.exists(ls_path_mnt) THEN                                                 #170406-00030#1 mark
      IF os.Path.exists(ls_path_erp) THEN                                                  #170406-00030#1 add
         #若有設定直接送印，則樣板路徑抓直接送印設定的變數
#         LET ls_path = cl_gr_get_4rpdir(g_rep_code,g_gzgd[l_ac].gzgd003,g_rlang,"W")      #150901-00021#1 mark
         IF g_erpid = "T100ERP" THEN                                                       #160608-00012#1 add
            LET ls_path = cl_rpt_get_4rpdir(g_rep_code,g_gzgd[l_ac].gzgd003,g_rlang,"W")   #150901-00021#1 add
         ELSE                                                                              #160608-00012#1 add
            LET ls_path = cl_rpt_get_4rpdir(g_rep_code,g_gzgd[l_ac].gzgd003,g_rlang,"L")   #160608-00012#1 add
         END IF                                                                            #160608-00012#1 add         
         LET g_4rp_dir = ls_path                                                           #170406-00030#1 add
         LET ls_path = os.Path.join(ls_path,lc_gzgd007||".4rp")
         DISPLAY 'ls_path: ', ls_path
         IF fgl_report_loadCurrentSettings(ls_path) THEN                                  #170406-00030#1 mark  #170912-00018#1 remark
            #191127-00011#2 add(s)
            CALL cl_rpt_get_date_fmt() RETURNING g_date_fmt
            SELECT ooef023 INTO l_ooef023
              FROM ooef_t
             WHERE ooefent = g_enterprise AND ooef001 = g_site AND ooefstus = "Y"
            #191127-00011#2 add(e)
#         IF fgl_report_loadCurrentSettings(lc_gzgd007||".4rp") THEN                        #170406-00030#1 add #170912-00018#1 mark
            #191127-00011#2 mark(s)
#            #根據組織基本資料維護作業(aooi100)取得日期格式和數字格式      
#            SELECT ooef013, ooef023 INTO l_ooef013, l_ooef023
#              FROM ooef_t
#              WHERE ooefent = g_enterprise AND ooef001 = g_site AND ooefstus = "Y" 
#     
#            #160817-00023#1 add(s)
#            #日期格式的SCC碼是10
#            SELECT gzcbl004 INTO l_gzcbl004 FROM gzcbl_t WHERE gzcbl001 = 10 AND gzcbl002 = l_ooef013 AND gzcbl003 = g_lang
#            LET g_date_fmt = l_gzcbl004
#            IF cl_null(g_date_fmt) THEN LET g_date_fmt = "yyyy/mm/dd" END IF
#            #160817-00023#1 add(e)
            #191127-00011#2 mark(e)            
            #160817-00023#1 mark(s)
#            CASE l_ooef013
#               WHEN "1"
#                  LET l_date_fmt = "Y4MD/"
#                  LET g_date_fmt = "yyyy/mm/dd"
#               WHEN "2"
#                  LET l_date_fmt = "MDY4/"
#                  LET g_date_fmt = "mm/dd/yyyy"                    
#               WHEN "3"
#                  LET l_date_fmt = "DMY4/"
#                  LET g_date_fmt = "dd/mm/yyyy"                    
#               OTHERWISE   
#                  LET l_date_fmt = "Y4MD/"
#                  LET g_date_fmt = "yyyy/mm/dd"                    
#            END CASE
            #160817-00023#1 mark(e)
     
            CASE l_ooef023
               WHEN "1"   #範例 ###,###,##&.&&
                  LET l_num_fmt_all = "$:,:.:"
               WHEN "2"   #範例 ###.###.##&,&&
                  LET l_num_fmt_all = "$:.:,:"
            END CASE
     
            CALL fgl_report_configureLocalization(NULL, NULL, l_num_fmt_all, NULL)   
            CALL cl_gr_init_pageheader()
         END IF
     
         IF g_save_flag = "Y" THEN
            LET ls_outputformat = "2"        # 一律用PDF
            IF cl_null(g_savefilename) THEN  
               LET ls_savefilename = cl_gr_savename(g_rep_code,g_gzgd[l_ac].gzgd000) #設定歷史留存的檔名
               LET ls_savefilename1 = ls_savefilename
#               LET ls_savefilename = os.Path.join(os.Path.join(FGL_GETENV("PATH4RP"),"history"),ls_savefilename) #報表主機看的懂的路徑   #150916-00002#1 mark
               LET ls_savefilename = os.Path.join(os.Path.join(cl_rpt_get_env_global("PATH4RP"),"history"),ls_savefilename) #在報表主機上開啟的路徑   #150916-00002#1 mark
               LET g_savefilename = ls_savefilename   
     
               #create new svg file and chmod 777
               #先產生歷史報表同檔名的空檔案
#               LET l_svg_name=os.Path.join(os.Path.join(FGL_GETENV("MNT4RP"),"history"),ls_savefilename1) #AP主機看的懂的路徑   #150916-00002#1 mark
#               LET l_svg_name=os.Path.join(os.Path.join(cl_rpt_get_env_global("MNT4RP"),"history"),ls_savefilename1)            #在ERP主機上開啟的路徑   #150916-00002#1 add   #170725-00086#1 mark
               LET l_svg_name=os.Path.join(os.Path.join(cl_rpt_get_env_global("PATH4RP"),"history"),ls_savefilename1)            #在ERP主機上開啟的路徑   #150916-00002#1 add    #170725-00086#1 add
#               IF NOT os.Path.exists(l_svg_name) THEN   #170725-00086#1 mark
                  LET l_svg_file=base.channel.create()
                  CALL l_svg_file.setdelimiter("")
                  CALL l_svg_file.openfile(l_svg_name CLIPPED, "w" )
                  CALL l_svg_file.close()                      
#               END IF   #170725-00086#1 mark 
#               IF os.Path.chrwx(l_svg_name,511) THEN END IF  #修改權限   #170725-00086#1 mark              
            END IF 
            #檔案輸出前先刪除舊檔
            CALL os.Path.delete(g_savefilename) RETURNING l_delete
            IF l_delete = 1 THEN
               DISPLAY 'delete from:', g_savefilename
            END IF
                 
            CALL fgl_report_setOutputFileName(ls_savefilename)
         END IF 
     
         #根據報表id呼叫Layout Editor
         IF (l_action = 'layout_edit') THEN #若按下按鈕
            SELECT gzgd000 INTO lc_gzgd000 FROM gzgd_t WHERE gzgdstus = "Y" AND gzgd001 = g_rep_code
                                                         AND gzgd002 = g_gzgd[l_ac].gzgd002
                                                         AND gzgd003 = g_gzgd[l_ac].gzgd003
                                                         AND gzgd007 = g_gzgd[l_ac].gzgd007
            #CALL cl_gr_edit(lc_gzgd000) #呼叫Layout Editor
         ELSE
            CASE ls_outputformat
               WHEN "9"
                  CALL cl_gr_mail()
               OTHERWISE
                  #取得報表元件代號說明(gzdel003)
                  #LET l_str_gzdel = cl_gr_get_progname(g_rep_code,g_dlang,2)  #160921-00007#3  mark                   
                  #LET l_str_gzdel = cl_get_progname(g_prog,g_dlang,2)          #160921-00007#3  add  #161227-00007#1 mark
                  #161227-00007#1 add -(s)
                  IF NOT cl_null(g_rep_code_desc) THEN
                        LET l_str_gzdel = g_rep_code_desc
                  ELSE
                        LET l_str_gzdel = cl_gr_get_progname(g_rep_code,g_dlang,2)        
                  END IF
                  #161227-00007#1 add -(e)
                  #轉換直接送印(azzi320)傳過來的代碼
                  IF g_dp_flag = 'Y' THEN
                     CASE g_dp.gzgn005
                        WHEN "P"
                           LET ls_outputformat = 2
                        WHEN "X"   #XLS-one page
                           LET ls_outputformat = 3
                        WHEN "F"   #XLS-分成多頁
                           LET ls_outputformat = 4
                        WHEN "B"   #XLSX-one page
                           LET ls_outputformat = 5
                        WHEN "E"   #XLSX-分成多頁
                           LET ls_outputformat = 6
                        WHEN "D"   #Word
                           LET ls_outputformat = 7
                        WHEN "I"
                           LET ls_outputformat = 8
                        WHEN "L"
                           LET ls_outputformat = "L"
                        #150921-00060#1 add(s)
                        WHEN "S"
                           LET ls_outputformat = "S"
                        #150921-00060#1 add(e)
                        WHEN "G"                       #170105-00009#2 add
                           LET ls_outputformat = "1"   #170105-00009#2 add                     
                     END CASE
                  END IF
                  
                  #171206-00020#1 add(s)
                  IF NOT cl_null(g_rep_file_format) AND g_rep_file_format MATCHES "[1-8LS]" THEN
                     LET ls_outputformat = g_rep_file_format
                  END IF
                  #171206-00020#1 add(e)
                  
                  CALL cl_gr_get_gre_version() RETURNING l_gre_ver    #171211-00055#1 add
                  
                  CASE ls_outputformat
                     WHEN "1"
                        &ifdef v1   #181213-00050#1 add
                          LET ls_output = "SVG"
                        &endif       #181213-00050#1 add
                        
                        #181213-00050#1 add(s)
                        &ifndef v1            
                           #透過GBC運行底層抓到是GWC，故留著GWC判斷                        
#                           IF FGL_GETENV("GUIMODE") = "GWC" OR FGL_GETENV("GUIMODE") = "GBC" THEN   #190910-00011 mark
                           IF l_frontend <> "GDC"  THEN   #190910-00011 add
                              LET g_err_msg = "NOT isGDC, GUIMODE:",FGL_GETENV("GUIMODE")
                              DISPLAY g_err_msg
                              CALL cl_gr_logfile()
                              LET ls_output = "Browser"                        
                           ELSE
                              LET g_err_msg = "isGDC"
                              CALL cl_gr_logfile()
                              LET ls_output = "SVG"
                           END IF  
                        &endif
                        #181213-00050#1 add(e)
                        
                        CALL fgl_report_setTitle(l_str_gzdel)   #匯出是SVG格式才執行此API
                        CALL fgl_report_configureSVGDevice(NULL,NULL,FALSE,NULL) #改善列印品質
                        #CALL fgl_report_setSVGCompression(TRUE)  #壓縮svg檔案  #改善列印空白議題
                        IF NOT cl_null(g_print_cnt_sql) THEN
                           CALL cl_gr_print_cnt(1) #回寫列印次數
                        END IF
                        #判斷是否為中一刀
#                        IF cl_gr_chk_landscape_paper(ls_path_mnt) THEN                           #170406-00030#1 mark
                        IF cl_gr_chk_landscape_paper(ls_path_erp) THEN                            #170406-00030#1 add
                           CALL fgl_report_setSVGOrientationRequested("portrait")
                        END IF
                     WHEN "2"
                        LET ls_output = "PDF"
                     WHEN "3" #合併one page
                        LET ls_output = "XLS"
#                        CALL fgl_report_configureXLSDevice(1,100000, true,true,true,true,TRUE)   #150930-00022#3 mark
#                        CALL fgl_report_configureXLSDevice(1,100000,false,true,true,true,TRUE)    #150930-00022#3 add   #171211-00055#1 mark
                        #171211-00055#1 add(s)
                        IF l_gre_ver >= "2.50.29" THEN
                           CALL fgl_report_configureXLSDevice(1,100000,false,false,false,true,TRUE)
                        ELSE
                           CALL fgl_report_configureXLSDevice(1,100000,false,true,true,true,TRUE)
                        END IF
                        #171211-00055#1 add(e)
                     WHEN "4" #分成多頁
                        LET ls_output = "XLS"
#                        CALL fgl_report_configureXLSDevice(1,100000,false,true,true,true,false)   #150930-00022#3 add #171211-00055#1 mark
                        #171211-00055#1 add(s)
                        IF l_gre_ver >= "2.50.29" THEN
                           CALL fgl_report_configureXLSDevice(1,100000,false,false,false,true,false)
                        ELSE
                           CALL fgl_report_configureXLSDevice(1,100000,false,true,true,true,false)
                        END IF
                        #171211-00055#1 add(e)
                     WHEN "5" #合併one page
                        LET ls_output = "XLSX"
#                        CALL fgl_report_configureXLSXDevice(1,100000, true,true,true,true,TRUE)  #150930-00022#3 mark
#                        CALL fgl_report_configureXLSXDevice(1,100000,false,true,true,true,TRUE)   #150930-00022#3 add #171211-00055#1 mark
                        #171211-00055#1 add(s)
                        IF l_gre_ver >= "2.50.29" THEN
                           CALL fgl_report_configureXLSXDevice(1,100000,false,false,false,true,TRUE)
                        ELSE
                           CALL fgl_report_configureXLSXDevice(1,100000,false,true,true,true,TRUE)
                        END IF
                        #171211-00055#1 add(e)
                     WHEN "6" #分成多頁
                        LET ls_output = "XLSX"
#                        CALL fgl_report_configureXLSXDevice(1,100000,false,true,true,true,false)  #150930-00022#3 add #171211-00055#1 mark
                        #171211-00055#1 add(s)
                        IF l_gre_ver >= "2.50.29" THEN
                           CALL fgl_report_configureXLSXDevice(1,100000,false,false,false,true,false)
                        ELSE
                           CALL fgl_report_configureXLSXDevice(1,100000,false,true,true,true,false)
                        END IF
                        #171211-00055#1 add(e)                        
                     WHEN "7" #WORD
                         LET ls_output = "RTF"
                         CALL fgl_report_setTitle(l_str_gzdel)
                     WHEN "8"
                        LET ls_output = "HTML"
                     WHEN "L" #輸出種類為Local Print
                        IF NOT cl_null(g_print_cnt_sql) THEN
                           CALL cl_gr_print_cnt(g_dp.gzgn007)    #回寫列印次數
                        END IF   
                        IF g_dp.gzgn009= "Y" THEN                        #立即列印否
                           CALL fgl_report_configureSVGPreview("PrintOnNamedPrinter")
                        ELSE
                           CALL fgl_report_configureSVGPreview("ShowPrintDialog")
                        END IF
                        CALL fgl_report_setSVGCopies(g_dp.gzgn007)       #設定列印份數
                        IF NOT cl_null(g_dp.gzgn006) THEN                #若有指定印表機則CALL API，若無則自動預設系統印表機
                           CALL fgl_report_setSVGPrinterName(g_dp.gzgn006)
                        END IF
#                        IF cl_gr_chk_landscape_paper(ls_path) THEN              #中一刀      #160912-00001#1-1 mark
#                        IF cl_gr_chk_landscape_paper(ls_path_mnt) THEN           #中一刀      #160912-00001#1-1 add   #170406-00030#1 mark
                        IF cl_gr_chk_landscape_paper(ls_path_erp) THEN           #中一刀      #160912-00001#1-1 add    #170406-00030#1 add
                           CALL fgl_report_setSVGOrientationRequested("portrait")
                        END IF
                        CALL fgl_report_configureSVGDevice(NULL,NULL,FALSE,NULL) #改善列印品質
                        CALL fgl_report_setSVGCompression(TRUE)                  #壓縮svg檔案
                     #150921-00060#1 add(s)
                     WHEN "S"
                        IF NOT cl_null(g_print_cnt_sql) THEN
                           CALL cl_gr_print_cnt(g_dp.gzgn007)    #回寫列印次數
                        END IF   
                     
                        LET ls_output = "Printer"
                        LET li_preview = 0
                        
                        IF NOT cl_null(g_dp.gzgn006) THEN #若有指定印表機則CALL API，若無則自動預設系統印表機
                           CALL fgl_report_setPrinterName(g_dp.gzgn006)
                        END IF
                        CALL fgl_report_setSVGCopies(g_dp.gzgn007)               #設定列印份數
#                        IF cl_gr_chk_landscape_paper(ls_path) THEN              #中一刀   #160912-00001#1-1 mark
#                        IF cl_gr_chk_landscape_paper(ls_path_mnt) THEN           #中一刀   #160912-00001#1-1 add   #170406-00030#1 mark
                        IF cl_gr_chk_landscape_paper(ls_path_erp) THEN           #中一刀   #160912-00001#1-1 add    #170406-00030#1 add
#                           CALL fgl_report_setSVGOrientationRequested("portrait")      #170828-00002#1 mark   
                           CALL fgl_report_setPrinterOrientationRequested("portrait")   #170828-00002#1 add
                        END IF                       
#                        LET INT_FLAG = TRUE    #151201-00015#1 mark
                        LET g_rep_flag = TRUE   #151201-00015#1 add
                     #150921-00060#1 add(e)    

                     #170815-00008#1 add(s)
                     WHEN "10"
                        LET l_sampledata_flag = "Y"
                        LET ls_outputformat = "1"
                        LET ls_output = "SVG"
                        LET li_preview = 0
                        CALL fgl_report_configureSVGDevice(NULL,NULL,FALSE,NULL) #改善列印品質
                        CALL fgl_report_setSVGCompression(TRUE)  #壓縮svg檔案                       
                     #170815-00008#1 add(e)
                        #選擇下載sampledata需進第二次才能跳出
                        LET g_rep_flag = NULL   #190321-00031 add
                  END CASE

                #181213-00050#1 add(s)
                &ifndef v1
                IF ls_output = "Browser" THEN
                     LET l_grehttpd_dir = cl_rpt_get_env_global("WINGREHTTPDROOT")   #190408-00033#2 modify
#                     LET l_grehttpd_dir = "D:/T100_gr/grehttpd/"   #180525寫死待改   #190408-00033#2 mark
                     LET gre_private_dir = fgl_getenv("GRE_PRIVATE_R_DIR")
                     LET gre_public_dir = fgl_getenv("GRE_PUBLIC_R_DIR")
                     LET g_err_msg = "gre_private_dir:",gre_private_dir,"  gre_public_dir:",gre_public_dir
                     CALL cl_gr_logfile()

                     LET l_uuid=security.RandomGenerator.CreateUUIDString()

                     -- Display variables
                     DISPLAY "DEBUG: GRE_PRIVATE_R_DIR = ", gre_private_dir
                     DISPLAY "DEBUG: GRE_PUBLIC_R_DIR = ", gre_public_dir
                     DISPLAY "DEBUG: GRE_PRIVATE_R_URL_PREFIX = ", fgl_getenv("GRE_PRIVATE_R_URL_PREFIX")
                     DISPLAY "DEBUG: GRE_PUBLIC_R_URL_PREFIX = ", fgl_getenv("GRE_PUBLIC_R_URL_PREFIX")
                     DISPLAY "DEBUG: GRE_REPORT_VIEWER_R_URL_PREFIX = ", fgl_getenv("GRE_REPORT_VIEWER_R_URL_PREFIX")
                     #GRE_PRIVATE_R_URL_PREFIX範例http://IP:80/wt30dev/ua/report-r/private/c10f91ed04f2ee70448d6d13ae66e217
                     LET g_err_msg = "GRE_PRIVATE_R_URL_PREFIX:",fgl_getenv("GRE_PRIVATE_R_URL_PREFIX"),"  GRE_PUBLIC_R_URL_PREFIX:",fgl_getenv("GRE_PUBLIC_R_URL_PREFIX"),"GRE_REPORT_VIEWER_R_URL_PREFIX:",fgl_getenv("GRE_REPORT_VIEWER_R_URL_PREFIX")
                     CALL cl_gr_logfile()

                     RUN "chmod -R 777 " || fgl_getenv("GRE_PRIVATE_R_DIR") || "/.."
                     #grehttpd產生檔案的路徑
                     CALL fgl_report_setBrowserDocumentDirectory(l_grehttpd_dir||gre_private_dir||"/"||l_uuid)   
                     #報表字型的路徑，會放在D:T100_gr/grehttpd/public/fonts 放在公用目錄為了不用重產字型，效能佳
                     CALL fgl_report_setBrowserFontDirectory(l_grehttpd_dir||gre_public_dir||"/fonts")  
                     #grehttpd文件url路徑  例http://IP:80/wt30dev/ua/report-r/private/c10f91ed04f2ee70448d6d13ae66e217
                     CALL fgl_report_setBrowserDocumentDirectoryURL(fgl_getenv("GRE_PRIVATE_R_URL_PREFIX")||"/"||l_uuid)
                      #grehttpd字型url路徑 
                     CALL fgl_report_setBrowserFontDirectoryURL(fgl_getenv("GRE_PUBLIC_R_URL_PREFIX")||"/fonts")

                     -- Open the URL:
                     #GRE_REPORT_VIEWER_URL_PREFIX範例http://IP/wt30dev/ua/report/viewer
                     DISPLAY "DEBUG: URL = ", fgl_getenv("GRE_REPORT_VIEWER_URL_PREFIX")||"/viewer.html?reportId="||l_uuid||"&privateUrlPrefix="||fgl_getenv("GRE_PRIVATE_R_URL_PREFIX")
                     LET g_err_msg = "DEBUG: URL=",fgl_getenv("GRE_REPORT_VIEWER_URL_PREFIX")||"/viewer.html?reportId="||l_uuid||"&privateUrlPrefix="||fgl_getenv("GRE_PRIVATE_R_URL_PREFIX")

                     CALL cl_gr_logfile()
                     
                     #URL組成範例http://IP/wt30dev/ua/report/viewer/viewer.html?reportId=4277768D-C9F8-4106-BC05-C71C4F1317ED&privateUrlPrefix=http://IP:80/wt30dev/ua/report-r/private/c10f91ed04f2ee70448d6d13ae66e217
                     #l_uuid = "4277768D-C9F8-4106-BC05-C71C4F1317ED"
                     #260327-00001#4 add(s)
                     IF FGL_GETENV("T100NANAWEB") = 'Y' THEN
                        DISPLAY "T100NaNaWeb Run GR Report"
                        #回傳網址處理
                        LET la_param.prog = "adzp902"
                        LET la_param.param[1] = "open_t100_report_gr"
                        LET la_param.param[2] = fgl_getenv("GRE_REPORT_VIEWER_URL_PREFIX")||"/viewer.html?reportId="||l_uuid||"&privateUrlPrefix="||fgl_getenv("GRE_PRIVATE_R_URL_PREFIX")
                        LET lc_str = util.JSON.stringify( la_param )   
                        CALL cl_cmdrun(lc_str)
                     ELSE  
                        #直接開啟
                        CALL ui.Interface.frontCall( "standard", "launchurl", [fgl_getenv("GRE_REPORT_VIEWER_URL_PREFIX")||"/viewer.html?reportId="||l_uuid||"&privateUrlPrefix="||fgl_getenv("GRE_PRIVATE_R_URL_PREFIX")], [] )
                     END IF
                     #260327-00001#4 add(e)
                     #CALL ui.Interface.frontCall( "standard", "launchurl", [fgl_getenv("GRE_REPORT_VIEWER_URL_PREFIX")||"/viewer.html?reportId="||l_uuid||"&privateUrlPrefix="||fgl_getenv("GRE_PRIVATE_R_URL_PREFIX")], [] ) #260327-00001#4 mark
                  END IF 
                  &endif
                  #181213-00050#1 add(e)  
                  
                  CALL fgl_report_selectDevice(ls_output)
                  CALL fgl_report_selectPreview(li_preview)
                  IF g_sendmail.getLength() > 0 THEN
                     LET g_gr_filename = cl_gr_jmail_filename()
#                     LET g_mail_attach = os.Path.join(os.Path.join(FGL_GETENV("MNT4RP"),"tmp"),g_gr_filename)             #150916-00002#1 mark
#                     LET g_mail_attach = os.Path.join(os.Path.join(cl_rpt_get_env_global("MNT4RP"),"tmp"),g_gr_filename)   #150916-00002#1 add   #170725-00086#1 mark
                     LET g_mail_attach = os.Path.join(os.Path.join(cl_rpt_get_env_global("PATH4RP"),"tmp"),g_gr_filename)   #150916-00002#1 add   #170725-00086#1 add
                     LET g_mail_attach_win=""
#                     LET l_drive_str = os.Path.join(fgl_getenv("PATH4RP"),"tmp")                                          #150916-00002#1 mark
                     LET l_drive_str = os.Path.join(cl_rpt_get_env_global("PATH4RP"),"tmp")                                #150916-00002#1 add
                     LET g_mail_attach_win =os.Path.join(l_drive_str,g_gr_filename)
                     CALL fgl_report_setOutputFileName(g_mail_attach_win)
                  ELSE
                     LET g_mail_attach = NULL
                  END IF
          
                  IF l_sampledata_flag = "Y" THEN
                     #170725-00086#1 mark(s)
#                     IF g_view_cnt > 50 THEN
#                        LET g_gr_progress = 'Y'
#                        CALL cl_progress_bar(2)
#                     END IF
                     #170725-00086#1 mark(e)
                     LET handler = cl_gr_sampledata(ls_dir,g_gzgd[l_ac].gzgd007)
                     LET g_sampledata_flag = "Y"   #170808-00030#1 add
                     LET g_sampledata_ac = l_ac                     
                     #170808-00030#1 mark(s)
#                     INITIALIZE g_errparam TO NULL
#                     LET g_errparam.code = 'azz-00677'
#                     LET g_errparam.popup = TRUE
#                     LET g_errparam.replace[1] = "\n",g_sampledata_path                                                    #160608-00012#1 add
#                     CALL cl_err()
                     #170808-00030#1 mark(e)
                  ELSE
                     #250912-00022#1 add(s)
                     IF FGL_GETENV("ATHENA") == "1" THEN
                        CALL fgl_report_selectPreview(FALSE)
                        CALL fgl_report_setOutputFileName(os.Path.join(os.Path.join(cl_rpt_get_env_global("PATH4RP"),"tmp"),FGL_GETENV("ATHENA_FILE")))
                     END IF
                     #250912-00022#1 add(e)
                     
                     #260327-00001#4 add(s)
                     LET l_ostype = FGL_GETENV("ostype")
                     IF NOT ((FGL_GETENV("GUIMODE") != "GWC" OR FGL_GETENV("GUIMODE") IS NULL) AND l_ostype.toUpperCase() = "WINDOWS") THEN
                        #IF FGL_GETENV("T100NANAWEB") = 'Y' AND ls_outputformat != "1" THEN                         #260430-00024#1 mark
                        #260430-00024#1 add(s)
                        IF     FGL_GETENV("T100NANAWEB") = 'Y'
                           AND ls_outputformat != "1"
                           AND g_save_flag != "Y"
                           THEN
                        #260430-00024#1 add(e)
                           IF g_sendmail.getLength() > 0 AND g_mail_cnt = g_sendmail.getLength() THEN
                              DISPLAY "T100NANAWEB Send Mail !!"
                           ELSE
                              LET ls_filename = cl_gr_jmail_filename()
                              CASE ls_outputformat
                                 WHEN "2"
                                    LET ls_filename = ls_filename,".pdf"
                                 WHEN "3"
                                    LET ls_filename = ls_filename,".xls"
                                 WHEN "4"
                                    LET ls_filename = ls_filename,".xls"
                                 WHEN "5"
                                    LET ls_filename = ls_filename,".xlsx"
                                 WHEN "6"
                                    LET ls_filename = ls_filename,".xlsx"
                                 WHEN "7"
                                    LET ls_filename = ls_filename,".rtf"
                                 WHEN "8"
                                    LET ls_filename = ls_filename,".html"
                              END CASE

                              CALL fgl_report_selectPreview(FALSE)
                              CALL fgl_report_setOutputFileName(os.Path.join(os.Path.join(cl_rpt_get_env_global("PATH4RP"),"tmp"),ls_filename))

                              LET l_url = "http://",cl_rpt_get_env_global("GRIP"),"/",FGL_GETENV("ZONE"),"/temp/gr/",ls_filename
                              
                              DISPLAY "T100NaNaWeb Run GR Report Export"
                              #回傳網址處理
                              LET la_param.prog = "adzp902"
                              LET la_param.param[1] = "open_t100_report_gr_export"
                              LET la_param.param[2] = l_url
                              LET la_param.param[3] = os.Path.join(os.Path.join(os.Path.join("/T100_gr",FGL_GETENV("ZONE")),"tmp"),ls_filename)                      
                              LET lc_str = util.JSON.stringify( la_param )   
                              CALL cl_cmdrun(lc_str)
                           END IF
                        END IF
                     END IF
                     #260327-00001#4 add(e)
                  
                     #路徑改抓動態
                     #組report location path/file name
                     CALL cl_gr_setDistributed(g_rep_code,g_gzgd[l_ac].gzgd003,lc_gzgd007)
	   	            LET handler = fgl_report_commitCurrentSettings()
                     IF (g_sendmail.getLength() > 0 AND g_mail_cnt >= g_sendmail.getLength()) OR g_sendmail.getLength() = 0 THEN
                        LET l_save = ''
                        CALL cl_gr_history_save() RETURNING l_save   #判斷是否需留存歷史報表
                        IF(l_save = 'Y') THEN                         #須留存
                           CALL cl_gr_history()
                        END IF
                     END IF
                     #LET g_rep_flag = TRUE #260327-00001#4 add   #260430-00024#1 mark
                  END IF
            END CASE #CASE ls_outputformat
         END IF   #IF (l_action = 'layout_edit') THEN #若按下按鈕
      ELSE
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'azz-00676'
         LET g_errparam.extend = ls_path
         LET g_errparam.popup = TRUE
         CALL cl_err()
      
#         EXIT PROGRAM -1   #150904-00005#4 mark
         CALL cl_log_ap_exiterror("-1","azz-00676")   #150904-00005#4 add
      END IF   #IF os.Path.exists(ls_path_mnt) 
   END IF   #IF INT_FLAG
   
#   IF g_dp_flag = 'Y' AND (g_save_flag = 'N' OR g_save_flag = 'S') THEN   #181213-00050#1 mark
   IF (g_dp_flag = 'Y' AND (g_save_flag = 'N' OR g_save_flag = 'S') OR     #181213-00050#1 add
       (l_frontend <> "GDC") AND g_gzgd.getLength()=1)  THEN   #190910-00011 add
#       (FGL_GETENV("GUIMODE")="GWC") AND g_gzgd.getLength()=1)  THEN       #181213-00050#1 add   #190910-00011 mark
#      LET INT_FLAG = TRUE    #151201-00015#1 mark
      LET g_rep_flag = TRUE   #151201-00015#1 add
   END IF
   
   LET g_err_msg = "Finish! "   #160912-00001#1-2 add
   CALL cl_gr_logfile()         #160912-00001#1-2 add
   
   #190327-00044#1 add(s)
   IF l_site_t = 'ALL' THEN
      LET g_site = l_site_t
   END IF
   #190327-00044#1 add(e)
   
   RETURN handler
END FUNCTION

################################################################################
# Descriptions...: 將GR表頭公用變數從資料庫讀出
# Memo...........:
# Usage..........: CALL cl_gr_init_pageheader()
#                  
# Input parameter: none
# Return code....: void
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_init_pageheader()
   DEFINE l_ooefl004    LIKE ooefl_t.ooefl004  #組織名稱
   DEFINE l_gzgh043     STRING                 #數值屬性的小數位數
   DEFINE l_success     LIKE type_t.num5       #成功與否
   DEFINE l_logopath    STRING
   DEFINE l_i,l_j       INTEGER
   DEFINE l_str         STRING
   DEFINE l_gzgm003     LIKE gzgm_t.gzgm003    #表頭代號 
   DEFINE l_gzgk006     LIKE gzgk_t.gzgk006    #顯示內容
   DEFINE l_ooek004     LIKE ooek_t.ooek004    #列印國家地區
   DEFINE l_ooek005     LIKE ooek_t.ooek005    #列印州/省/直轄市
   DEFINE l_gzgd014     LIKE gzgd_t.gzgd014    #表頭位置
   DEFINE l_prog_t      LIKE type_t.chr20      #作業編號舊值   #150709-00023#1 add
   DEFINE l_pri_con     LIKE gzgj_t.gzgj009    #150904-00005#5 add
   DEFINE l_sql         STRING                 #紀錄撈取SQL段用    #180524-00013#2 add
   DEFINE l_cnt         INTEGER                #紀錄COUNT 出來的值 #180524-00013#2 add   
   DEFINE l_site_t      STRING                 #190327-00044#1 add   #備份g_site用
   
   #190327-00044#1 add(s)
   LET l_site_t = ""
   IF g_site = "ALL" THEN
      LET l_site_t = g_site
      SELECT gzxa007 INTO g_site FROM gzxa_t
       WHERE gzxaent = g_enterprise AND gzxa001 = g_account
   END IF
   #190327-00044#1 add(e)
        
   #製表者
   LET g_user_name = cl_get_username(g_user)
   #製表日期
   LET g_pdate = cl_get_today()
   LET g_ptime = cl_get_time()

   #判斷是否設定BPM簽核
   IF NOT cl_null(g_rep_docno) THEN 
      LET l_prog_t = g_prog               #150709-00023#1 add  
      LET g_prog = g_parentprog           #150709-00023#1 add #BPM判斷前先置換g_prog
      IF cl_bpm_chk_flow(g_rep_docno) THEN
         LET g_grPageHeader.sign_type = "b"   #簽核圖檔
      END IF
      LET g_prog = l_prog_t               #150709-00023#1 add #還原g_prog
   ELSE
      LET g_grPageHeader.sign_type = "t"  #簽核字串
   END IF

#   CALL cl_gr_get_numfmt_str() RETURNING l_gzgh043 #數值欄位格式   #150901-00021#1 mark
   CALL cl_rpt_get_numfmt_str() RETURNING l_gzgh043 #數值欄位格式   #150901-00021#1 add 

#   LET g_grPageHeader.pri_con = "N"   #列印條件顯示否   #150904-00005#5 mark

   #表頭列印位置,1:每頁顯示,2:只有第一頁顯示
   SELECT gzgd014 INTO l_gzgd014 FROM gzgd_t 
    WHERE gzgd000 = g_gzgd000
     AND gzgdstus = "Y"
   IF cl_null(l_gzgd014) THEN
      LET l_gzgd014 = "1"
   END IF
   LET g_grPageHeader.head_loc= l_gzgd014
   
   #取得表頭代號
   SELECT MAX(gzgm003) INTO l_gzgm003 FROM gzgm_t
    WHERE gzgmstus = "Y" AND gzgm001 = g_rep_code AND gzgm002 = g_rlang    #180131-00079#1 add
#    WHERE gzgmstus = "Y" AND gzgm001 = g_rep_code AND gzgm002 = g_dlang   #180131-00079#1 mark

   #抓不到表頭代號，抓YY
#   IF cl_null(l_gzgm003) THEN LET l_gzgm003 = "YY" END IF   #180524-00013#2 mark

   #180524-00013#2 add(s)
   #組好SQL指令
   LET l_sql = "SELECT COUNT(1) FROM gzgj_t WHERE gzgj001 = 'T100_YY'"
   #準備CURSOR
   PREPARE cl_gr_init_curs FROM l_sql
   #執行CURSOR
   EXECUTE cl_gr_init_curs INTO l_cnt

   #判斷如果l_gzgm003 是空值就往下執行
   IF cl_null(l_gzgm003) THEN
      #判斷COUNT不等於0就帶入T100_YY，若等於0就帶入YY
      IF l_cnt <> 0 THEN
         LET l_gzgm003 = "T100_YY"
      ELSE
         LET l_gzgm003 = "YY"
      END IF
   END IF
   #180524-00013#2 add(e)

   #公司logo
   SELECT gzgj009 INTO g_grPageHeader.logoPos FROM gzgj_t WHERE gzgj001 = l_gzgm003 AND gzgj002 = g_dlang
      AND gzgj010 = '1'   #150611-00032#3 add
   IF cl_null(g_grPageHeader.logoPos) THEN LET g_grPageHeader.logoPos = "0"  END IF #logo圖檔位置
   LET g_grPageHeader.logoWid = 2.5 #logo先寫死寬度
   LOCATE g_grPageHeader.logo IN MEMORY 
   SELECT ooek003,ooek004,ooek005 INTO g_grPageHeader.logo,l_ooek004,l_ooek005 FROM ooek_t
    WHERE ooekstus = "Y" AND ooekent = g_enterprise AND ooek001 = g_site AND ooek002 = g_dlang

   IF g_grPageHeader.logo IS NULL THEN    #240515-00001#1 remark    #240929-00003#1 unmark
   #IF cl_null(g_grPageHeader.logo) THEN    #240515-00001#1 add      #240929-00003#1 mark 
      #預設"no logo"路徑,當組織基本資料維護作業(aooi100)未設定logo時則顯示"no logo"
      LET l_logopath = os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("TOP"),"www"),"pic"),"reptpic")  
      LET l_logopath = os.Path.join(l_logopath,"no_logo.png")
      LOCATE g_grPageHeader.logo IN FILE l_logopath
      #160608-00012#1 add(s)
      IF g_grPageHeader.logo IS NULL THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'azz-00146'
         LET g_errparam.extend = ""
         LET g_errparam.popup = TRUE
         CALL cl_err()   
         
         LET g_rep_flag = TRUE
         RETURN
#         LET handler = NULL
#         RETURN handler
      END IF   
      #160608-00012#1 add(e)   
   END IF
   
   #150904-00005#5 add(s)
   #列印條件顯示否
   SELECT gzgj009 INTO l_pri_con FROM gzgj_t 
    WHERE gzgj001 = l_gzgm003 AND gzgj002 = g_dlang AND gzgj010 = '2'
   LET g_grPageHeader.pri_con = l_pri_con
   IF cl_null(g_grPageHeader.pri_con) THEN LET g_grPageHeader.pri_con = 'N' END IF
   #150904-00005#5 add(e)
   
   IF cl_null(l_ooek004) THEN LET l_ooek004 = "Y" END IF #列印國家地區
   
   IF cl_null(l_ooek005) THEN LET l_ooek005 = "Y" END IF #列印州/省/直轄市

   CALL cl_gr_get_addr(l_gzgm003,l_ooek004,l_ooek005)    #取得地址

   #取得表頭顯示內容
   LET l_str = "SELECT gzgk006 FROM gzgk_t ",
               " WHERE gzgk001 = '",l_gzgm003,"' AND gzgk002 = '",g_dlang,"' AND gzgk003 = ? AND gzgk004 = ? "
               #PGS-11084 add -S-
               #PGS-11084 add -E-
               ," AND gzgk014 = '1' "   #150611-00032#3 add
   PREPARE cl_gr_gzgk006_pre FROM l_str

   IF g_grPageHeader.title0101 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "1","1"
      LET g_grPageHeader.title0101 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0102 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "1","2"
      LET g_grPageHeader.title0102 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0103 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "1","3"
      LET g_grPageHeader.title0103 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0104 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "1","4"
      LET g_grPageHeader.title0104 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0201 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "2","1"
      LET g_grPageHeader.title0201 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0202 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "2","2"
      LET g_grPageHeader.title0202 = cl_gr_title(l_gzgk006)   
   END IF
   
   IF g_grPageHeader.title0203 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "2","3"
      LET g_grPageHeader.title0203 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0204 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "2","4"
      LET g_grPageHeader.title0204 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0301 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "3","1"
      LET g_grPageHeader.title0301 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0302 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "3","2"
      LET g_grPageHeader.title0302 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0303 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "3","3"
      LET g_grPageHeader.title0303 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0304 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "3","4"
      LET g_grPageHeader.title0304 = cl_gr_title(l_gzgk006)   
   END IF
   
   IF g_grPageHeader.title0401 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "4","1"
      LET g_grPageHeader.title0401 = cl_gr_title(l_gzgk006)
   END IF
  
   IF g_grPageHeader.title0402 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "4","2"
      LET g_grPageHeader.title0402 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0403 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "4","3"
      LET g_grPageHeader.title0403 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0404 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "4","4"
      LET g_grPageHeader.title0404 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0501 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "5","1"
      LET g_grPageHeader.title0501 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0502 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "5","2"
      LET g_grPageHeader.title0502 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0503 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "5","3"
      LET g_grPageHeader.title0503 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0504 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "5","4"
      LET g_grPageHeader.title0504 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0601 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "6","1"
      LET g_grPageHeader.title0601 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0602 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "6","2"
      LET g_grPageHeader.title0602 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0603 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "6","3"
      LET g_grPageHeader.title0603 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0604 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "6","4"
      LET g_grPageHeader.title0604 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0701 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "7","1"
      LET g_grPageHeader.title0701 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0702 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "7","2"
      LET g_grPageHeader.title0702 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0703 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "7","3"
      LET g_grPageHeader.title0703 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0704 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "7","4"
      LET g_grPageHeader.title0704 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0801 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "8","1"
      LET g_grPageHeader.title0801 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0802 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "8","2"
      LET g_grPageHeader.title0802 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0803 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "8","3"
      LET g_grPageHeader.title0803 = cl_gr_title(l_gzgk006)   
   END IF

   IF g_grPageHeader.title0804 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "8","4"
      LET g_grPageHeader.title0804 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageHeader.title0901 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "9","1"
      LET g_grPageHeader.title0901 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0902 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "9","2"
      LET g_grPageHeader.title0902 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0903 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "9","3"
      LET g_grPageHeader.title0903 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title0904 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "9","4"
      LET g_grPageHeader.title0904 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title1001 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "10","1"
      LET g_grPageHeader.title1001 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title1002 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "10","2"
      LET g_grPageHeader.title1002 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title1003 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "10","3"
      LET g_grPageHeader.title1003 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageHeader.title1004 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_pre INTO l_gzgk006 USING "10","4"
      LET g_grPageHeader.title1004 = cl_gr_title(l_gzgk006)
   END IF  
   
#150611-00032#3 add(s) 
   #取得表尾顯示內容
   LET l_str = "SELECT gzgk006 FROM gzgk_t ",
               " WHERE gzgk001 = '",l_gzgm003,"' AND gzgk002 = '",g_dlang,"' AND gzgk003 = ? AND gzgk004 = ? "
               #PGS-11084 add -S-
               #PGS-11084 add -E-
               ," AND gzgk014 = '2' "
   PREPARE cl_gr_gzgk006_footer_pre FROM l_str

   IF g_grPageFooter.title0101 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "1","1"
      LET g_grPageFooter.title0101 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageFooter.title0102 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "1","2"
      LET g_grPageFooter.title0102 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageFooter.title0103 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "1","3"
      LET g_grPageFooter.title0103 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageFooter.title0104 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "1","4"
      LET g_grPageFooter.title0104 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageFooter.title0201 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "2","1"
      LET g_grPageFooter.title0201 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageFooter.title0202 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "2","2"
      LET g_grPageFooter.title0202 = cl_gr_title(l_gzgk006)   
   END IF
   
   IF g_grPageFooter.title0203 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "2","3"
      LET g_grPageFooter.title0203 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageFooter.title0204 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "2","4"
      LET g_grPageFooter.title0204 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageFooter.title0301 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "3","1"
      LET g_grPageFooter.title0301 = cl_gr_title(l_gzgk006)
   END IF
   
   IF g_grPageFooter.title0302 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "3","2"
      LET g_grPageFooter.title0302 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageFooter.title0303 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "3","3"
      LET g_grPageFooter.title0303 = cl_gr_title(l_gzgk006)
   END IF

   IF g_grPageFooter.title0304 IS NULL THEN
      LET l_gzgk006 = ""
      EXECUTE cl_gr_gzgk006_footer_pre INTO l_gzgk006 USING "3","4"
      LET g_grPageFooter.title0304 = cl_gr_title(l_gzgk006)   
   END IF   
#150611-00032#3 add(e) 

   #190327-00044#1 add(s)
   IF l_site_t = 'ALL' THEN
      LET g_site = l_site_t
   END IF
   #190327-00044#1 add(e)

END FUNCTION

################################################################################
# Descriptions...: 初始化列印簽核相關變數
# Memo...........:
# Usage..........: CALL cl_gr_init_apr(p_document)
#                  
# Input parameter: p_document
# Return code....: 
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_init_apr(p_document)
   DEFINE p_document  STRING
   DEFINE p_table      STRING
   DEFINE l_wsbd        DYNAMIC ARRAY OF RECORD
          wsbd003       LIKE wsbd_t.wsbd003,  #使用者代號
          wsbd008       LIKE wsbd_t.wsbd008,  #手寫簽名圖檔
          wsbd006       LIKE wsbd_t.wsbd006,  #簽核日期   #170223-00005#2 add
          wsbd007       LIKE wsbd_t.wsbd007,  #簽核時間   #170223-00005#2 add
          wsbd010       LIKE wsbd_t.wsbd010,  #簽核關卡名稱   #200420-00014#3 add
          wsbd011       LIKE wsbd_t.wsbd011,  #簽核代理人     #230831-00036#3 add
          title         LIKE oocql_t.oocql004   #職稱
          
          END RECORD
   DEFINE l_pk_key     STRING                 #單據pk欄位的key/value組合   
   DEFINE l_img_path   STRING
   ###
   DEFINE l_img_path01 STRING
   DEFINE l_img_path02 STRING
   DEFINE l_img_path03 STRING
   DEFINE l_img_path04 STRING
   DEFINE l_img_path05 STRING
   DEFINE l_img_path06 STRING
   DEFINE l_img_path07 STRING
   DEFINE l_img_path08 STRING
   DEFINE l_img_path09 STRING
   DEFINE l_img_path10 STRING
   DEFINE l_img_path11 STRING
   DEFINE l_img_path12 STRING
   ###
   DEFINE l_sql        STRING
   DEFINE l_str        STRING      
   DEFINE l_i          LIKE type_t.num5
   DEFINE l_success    LIKE type_t.num5        #檢查是否成功
   DEFINE l_name       LIKE oofa_t.oofa011     #全名 
   DEFINE l_wsbd004    LIKE wsbd_t.wsbd004     #161219-00003#1 add
   DEFINE l_bpm_dt     LIKE wset_t.wset002     #BPM簽核歷程設定 #170223-00005#2 add
   DEFINE l_bpm_order  LIKE wset_t.wset002     #BPM簽核排序     #171211-00031#1 add
   DEFINE l_status     LIKE type_t.num5        #230313-00030#2 add

   IF g_grPageHeader.sign_show = "N" THEN RETURN END IF
   IF g_grPageHeader.sign_type = "b" THEN
      LOCATE l_wsbd[1].wsbd008 IN MEMORY
      LOCATE l_wsbd[2].wsbd008 IN MEMORY
      LOCATE l_wsbd[3].wsbd008 IN MEMORY
      LOCATE l_wsbd[4].wsbd008 IN MEMORY
      LOCATE l_wsbd[5].wsbd008 IN MEMORY
      LOCATE l_wsbd[6].wsbd008 IN MEMORY
      LOCATE l_wsbd[7].wsbd008 IN MEMORY
      LOCATE l_wsbd[8].wsbd008 IN MEMORY
      LOCATE l_wsbd[9].wsbd008 IN MEMORY
      LOCATE l_wsbd[10].wsbd008 IN MEMORY
      LOCATE l_wsbd[11].wsbd008 IN MEMORY
      LOCATE l_wsbd[12].wsbd008 IN MEMORY

      LOCATE g_grSign.sign_img01 IN MEMORY
      LOCATE g_grSign.sign_img02 IN MEMORY
      LOCATE g_grSign.sign_img03 IN MEMORY
      LOCATE g_grSign.sign_img04 IN MEMORY
      LOCATE g_grSign.sign_img05 IN MEMORY
      LOCATE g_grSign.sign_img06 IN MEMORY
      LOCATE g_grSign.sign_img07 IN MEMORY
      LOCATE g_grSign.sign_img08 IN MEMORY
      LOCATE g_grSign.sign_img09 IN MEMORY
      LOCATE g_grSign.sign_img10 IN MEMORY
      LOCATE g_grSign.sign_img11 IN MEMORY
      LOCATE g_grSign.sign_img12 IN MEMORY   
   
      #取得單據pk欄位的key, value組合
      LET l_pk_key = cl_replace_str(g_doc_key," ","")
      LET l_status = cl_bpm_chk_flow_status(l_pk_key)  #230313-00030#2 add
      
      #取得簽核關卡資訊
      #161219-00003#1 add(s)
      #最後一關的「簽核結果wsbd004」不是 'A' 已核准，則不列印簽核歷程 (這張單所有關卡都不印)
      LET l_sql = "SELECT wsbd004 FROM wsbd_t ",
                  " WHERE wsbdent = ",g_enterprise," AND wsbd001 = '",l_pk_key, "' ",
                  " ORDER BY wsbd006 DESC, wsbd007 DESC "
      PREPARE cl_gr_init_apr_wsbd004_pr FROM l_sql
      DECLARE cl_gr_init_apr_wsbd004_cs CURSOR FOR cl_gr_init_apr_wsbd004_pr
      FOREACH cl_gr_init_apr_wsbd004_cs INTO l_wsbd004
         EXIT FOREACH            
      END FOREACH
      #161219-00003#1 add(e)
      #IF l_wsbd004 = 'A' THEN      #161219-00003#1 add #230313-00030#2 mark
      IF l_wsbd004 = 'A' AND l_status THEN              #230313-00030#2 add
         LET l_bpm_order = cl_aws_prod_para(g_enterprise,"bpm-0009")   #171211-00031#1 add
         #先取手寫簽名圖檔
#         LET l_sql = "SELECT wsbd003,wsbd008 FROM wsbd_t ",                                                 #170223-00005#2 mark
#         LET l_sql = "SELECT wsbd003,wsbd008,wsbd006,wsbd007 FROM wsbd_t ",                                  #170223-00005#2 add   #200420-00014#3 mark
#         LET l_sql = "SELECT wsbd003,wsbd008,wsbd006,wsbd007,wsbd010 FROM wsbd_t ",                          #200420-00014#3 add   #230831-00036#3 mark 
          LET l_sql = "SELECT wsbd003,wsbd008,wsbd006,wsbd007,wsbd010,wsbd011 FROM wsbd_t ",                  #230831-00036#3 add         
   #                  " WHERE wsbdent = ",g_enterprise," AND wsbd001 = '",l_pk_key, "'",                     #161219-00003#1 mark
                     " WHERE wsbdent = ",g_enterprise," AND wsbd001 = '",l_pk_key, "' AND wsbd004 = 'A' "    #171211-00031#1 add
                     #171211-00031#1 mark(s)
#                     " WHERE wsbdent = ",g_enterprise," AND wsbd001 = '",l_pk_key, "' AND wsbd004 = 'A' ",   #161219-00003#1 add
#                     " ORDER BY  wsbd006, wsbd007 " #根據簽核日期時間排序
                     #171211-00031#1 mark(e)
         #171211-00031#1 add(s)
         IF l_bpm_order = "1" THEN
            LET l_sql = l_sql, " ORDER BY wsbd006 DESC, wsbd007 DESC "
         ELSE
            LET l_sql = l_sql, " ORDER BY wsbd006, wsbd007 " #根據簽核日期時間排序
         END IF
         #171211-00031#1 add(e)
         
         #取得awsi100報表列印簽核歷程
         LET l_bpm_dt = cl_aws_prod_para(g_enterprise,"bpm-0008")   #200420-00014#3 add         
         PREPARE cl_gr_init_apr_bpm_pr FROM l_sql
         DECLARE cl_gr_init_apr_bpm_cs CURSOR FOR cl_gr_init_apr_bpm_pr
         LET l_i = 1
         FOREACH cl_gr_init_apr_bpm_cs INTO l_wsbd[l_i].wsbd003,l_wsbd[l_i].wsbd008
                                           ,l_wsbd[l_i].wsbd006,l_wsbd[l_i].wsbd007                          #170223-00005#2 add
                                           ,l_wsbd[l_i].wsbd010                                              #200420-00014#3 add
                                           ,l_wsbd[l_i].wsbd011                                              #230831-00036#3 add
            #取得簽核關卡
            #取得使用者編號對應的全名與職稱
            CALL s_employee_get_userdata(l_wsbd[l_i].wsbd003,g_rlang) RETURNING l_success,l_name,l_wsbd[l_i].title
            
            #200420-00014#3 add(s)
            #使用簽核關卡名稱
            IF l_bpm_dt MATCHES '[345]' AND NOT cl_null(l_wsbd[l_i].wsbd010) THEN
               LET l_wsbd[l_i].title = l_wsbd[l_i].wsbd010
            END IF
            #200420-00014#3 add(e)
       
            #該簽核關卡無職稱=>預設為"簽核者"
            IF cl_null(l_wsbd[l_i].title) THEN
               LET l_wsbd[l_i].title = cl_getmsg("lib-00248",g_lang) CLIPPED
            END IF
            
            #230831-00036#3 add(s)
            IF l_wsbd[l_i].wsbd011 = "Y" THEN
               LET l_wsbd[l_i].title =  cl_getmsg("lib-00663",g_lang) , l_wsbd[l_i].title
            END IF
            #230831-00036#3 add(e)
            
            #無手寫簽名，則抓取gzxx_t.gzxx002(個人簽名圖檔)
            IF l_wsbd[l_i].wsbd008 IS NULL THEN
               SELECT gzxx002 INTO l_wsbd[l_i].wsbd008 FROM gzxx_t
                WHERE gzxxent = g_enterprise
                  AND gzxx001 = l_wsbd[l_i].wsbd003
            END IF
       
            #無手寫簽名且未設定個人簽名圖檔,將使用者姓名轉成圖檔 
            IF l_wsbd[l_i].wsbd008 IS NULL THEN
               #161219-00003#1 add(s)
               IF cl_null(l_name) THEN
                  CALL cl_gr_mkimg(l_wsbd[l_i].wsbd003 CLIPPED, "jpg") RETURNING l_img_path
                  LOCATE l_wsbd[l_i].wsbd008 IN FILE l_img_path
               ELSE
               #161219-00003#1 add(e)
                  CALL cl_gr_mkimg(l_name CLIPPED, "jpg") RETURNING l_img_path
                  LOCATE l_wsbd[l_i].wsbd008 IN FILE l_img_path
               END IF   #161219-00003#1 add
            END IF
            
            LET l_i = l_i + 1
         END FOREACH
      END IF   #161219-00003#1 add
      
      #職稱
      LET g_grSign.apr_str01 = l_wsbd[1].title
      LET g_grSign.apr_str02 = l_wsbd[2].title
      LET g_grSign.apr_str03 = l_wsbd[3].title
      LET g_grSign.apr_str04 = l_wsbd[4].title
      LET g_grSign.apr_str05 = l_wsbd[5].title
      LET g_grSign.apr_str06 = l_wsbd[6].title
      LET g_grSign.apr_str07 = l_wsbd[7].title
      LET g_grSign.apr_str08 = l_wsbd[8].title
      LET g_grSign.apr_str09 = l_wsbd[9].title
      LET g_grSign.apr_str10 = l_wsbd[10].title
      LET g_grSign.apr_str11 = l_wsbd[11].title
      LET g_grSign.apr_str12 = l_wsbd[12].title
      
      #170223-00005#2 add(s)
      #原有職稱+姓名，依據設定加入日期/時間
      LET l_bpm_dt = cl_aws_prod_para(g_enterprise,"bpm-0008")
      #200420-00014#3 mark(s)
#      CASE l_bpm_dt
#         WHEN '1'   #職稱+姓名+簽核日期
      #200420-00014#3 mark(e)
      #200420-00014#3 add(s)
      CASE
         WHEN l_bpm_dt = '1' OR l_bpm_dt = '4'  #1:職稱+姓名+簽核日期/4:關卡+姓名+簽核日期
      #200420-00014#3 add(e)
            LET g_grSign.apr_str01 = g_grSign.apr_str01,'\n',l_wsbd[1].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str02 = g_grSign.apr_str02,'\n',l_wsbd[2].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str03 = g_grSign.apr_str03,'\n',l_wsbd[3].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str04 = g_grSign.apr_str04,'\n',l_wsbd[4].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str05 = g_grSign.apr_str05,'\n',l_wsbd[5].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str06 = g_grSign.apr_str06,'\n',l_wsbd[6].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str07 = g_grSign.apr_str07,'\n',l_wsbd[7].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str08 = g_grSign.apr_str08,'\n',l_wsbd[8].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str09 = g_grSign.apr_str09,'\n',l_wsbd[9].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str10 = g_grSign.apr_str10,'\n',l_wsbd[10].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str11 = g_grSign.apr_str11,'\n',l_wsbd[11].wsbd006 USING g_date_fmt
            LET g_grSign.apr_str12 = g_grSign.apr_str12,'\n',l_wsbd[12].wsbd006 USING g_date_fmt

#         WHEN '2'   #職稱+姓名+簽核日期時間   #200420-00014#3 mark
         WHEN l_bpm_dt = '2' OR l_bpm_dt = '5'  #2:職稱+姓名+簽核日期時間/5:關卡+姓名+簽核日期時間   #200420-00014#3 add
            LET g_grSign.apr_str01 = g_grSign.apr_str01,'\n',l_wsbd[1].wsbd006 USING g_date_fmt,'\n',l_wsbd[1].wsbd007
            LET g_grSign.apr_str02 = g_grSign.apr_str02,'\n',l_wsbd[2].wsbd006 USING g_date_fmt,'\n',l_wsbd[2].wsbd007 
            LET g_grSign.apr_str03 = g_grSign.apr_str03,'\n',l_wsbd[3].wsbd006 USING g_date_fmt,'\n',l_wsbd[3].wsbd007 
            LET g_grSign.apr_str04 = g_grSign.apr_str04,'\n',l_wsbd[4].wsbd006 USING g_date_fmt,'\n',l_wsbd[4].wsbd007 
            LET g_grSign.apr_str05 = g_grSign.apr_str05,'\n',l_wsbd[5].wsbd006 USING g_date_fmt,'\n',l_wsbd[5].wsbd007 
            LET g_grSign.apr_str06 = g_grSign.apr_str06,'\n',l_wsbd[6].wsbd006 USING g_date_fmt,'\n',l_wsbd[6].wsbd007 
            LET g_grSign.apr_str07 = g_grSign.apr_str07,'\n',l_wsbd[7].wsbd006 USING g_date_fmt,'\n',l_wsbd[7].wsbd007 
            LET g_grSign.apr_str08 = g_grSign.apr_str08,'\n',l_wsbd[8].wsbd006 USING g_date_fmt,'\n',l_wsbd[8].wsbd007 
            LET g_grSign.apr_str09 = g_grSign.apr_str09,'\n',l_wsbd[9].wsbd006 USING g_date_fmt,'\n',l_wsbd[9].wsbd007 
            LET g_grSign.apr_str10 = g_grSign.apr_str10,'\n',l_wsbd[10].wsbd006 USING g_date_fmt,'\n',l_wsbd[10].wsbd007
            LET g_grSign.apr_str11 = g_grSign.apr_str11,'\n',l_wsbd[11].wsbd006 USING g_date_fmt,'\n',l_wsbd[11].wsbd007
            LET g_grSign.apr_str12 = g_grSign.apr_str12,'\n',l_wsbd[12].wsbd006 USING g_date_fmt,'\n',l_wsbd[12].wsbd007
      END CASE
      #170223-00005#2 add(e)
      
      LET l_img_path = os.Path.join(os.Path.join(FGL_GETENV("FGLASIP"),"pic"),"reptpic") 
      LET l_img_path = os.Path.join(l_img_path,"default.png")
      
      IF l_wsbd[1].wsbd008 IS NULL THEN
         LOCATE l_wsbd[1].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img01 = l_wsbd[1].wsbd008
      ELSE
         LET g_grSign.sign_img01 = l_wsbd[1].wsbd008
      END IF
      
      IF l_wsbd[2].wsbd008 IS NULL THEN
         LOCATE l_wsbd[2].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img02 = l_wsbd[2].wsbd008         
      ELSE
         LET g_grSign.sign_img02 = l_wsbd[2].wsbd008
      END IF

      IF l_wsbd[3].wsbd008 IS NULL THEN
         LOCATE l_wsbd[3].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img03 = l_wsbd[3].wsbd008         
      ELSE
         LET g_grSign.sign_img03 = l_wsbd[3].wsbd008
      END IF      
    
      IF l_wsbd[4].wsbd008 IS NULL THEN
         LOCATE l_wsbd[4].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img04 = l_wsbd[4].wsbd008         
      ELSE
         LET g_grSign.sign_img04 = l_wsbd[4].wsbd008
      END IF      

      IF l_wsbd[5].wsbd008 IS NULL THEN
         LOCATE l_wsbd[5].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img05 = l_wsbd[5].wsbd008         
      ELSE
         LET g_grSign.sign_img05 = l_wsbd[5].wsbd008
      END IF      
      
      IF l_wsbd[6].wsbd008 IS NULL THEN
         LOCATE l_wsbd[6].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img06 = l_wsbd[6].wsbd008         
      ELSE
         LET g_grSign.sign_img06 = l_wsbd[6].wsbd008
      END IF      

      IF l_wsbd[7].wsbd008 IS NULL THEN
         LOCATE l_wsbd[7].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img07 = l_wsbd[7].wsbd008         
      ELSE
         LET g_grSign.sign_img07 = l_wsbd[7].wsbd008
      END IF      

      IF l_wsbd[8].wsbd008 IS NULL THEN
         LOCATE l_wsbd[8].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img08 = l_wsbd[8].wsbd008         
      ELSE
         LET g_grSign.sign_img08 = l_wsbd[8].wsbd008
      END IF 

      IF l_wsbd[9].wsbd008 IS NULL THEN
         LOCATE l_wsbd[9].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img09 = l_wsbd[9].wsbd008         
      ELSE
         LET g_grSign.sign_img09 = l_wsbd[9].wsbd008
      END IF 

      IF l_wsbd[10].wsbd008 IS NULL THEN
         LOCATE l_wsbd[10].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img10 = l_wsbd[10].wsbd008         
      ELSE
         LET g_grSign.sign_img10 = l_wsbd[10].wsbd008
      END IF 

      IF l_wsbd[11].wsbd008 IS NULL THEN
         LOCATE l_wsbd[11].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img11 = l_wsbd[11].wsbd008         
      ELSE
         LET g_grSign.sign_img11 = l_wsbd[11].wsbd008
      END IF 

      IF l_wsbd[12].wsbd008 IS NULL THEN
         LOCATE l_wsbd[12].wsbd008 IN FILE l_img_path #將檔案內容讀取到BYTE變數
         LET g_grSign.sign_img12 = l_wsbd[12].wsbd008         
      ELSE
         LET g_grSign.sign_img12 = l_wsbd[12].wsbd008
      END IF       
   ELSE
      CALL cl_gr_tiptop_apr(p_document)  
    END IF
   
END FUNCTION

################################################################################
# Descriptions...: 取得樣板檔目錄
# Memo...........:
# Usage..........: LET ls_path = cl_gr_get_4rpdir(p_rep_code,p_gzgd003,p_lang,"L")
#                   
# Input parameter: p_rep_code  報表元件代號
#                : p_gzgd003   客製否
#                : p_lang      語言別
#                : p_os        系統別
# Return code....: ls_4rpdir   樣板檔路徑
#                : 
# Date & Author..: 2014 By 04010
# Modify.........: #150901-00021#1 改用cl_rpt_get_4rpdir
################################################################################
PUBLIC FUNCTION cl_gr_get_4rpdir(p_rep_code,p_gzgd003,p_lang,p_os)
   DEFINE p_rep_code   LIKE gzgf_t.gzgf001  #報表元件代號 
   DEFINE p_gzgd003    LIKE gzgd_t.gzgd003  #客製否
   DEFINE p_lang       LIKE gzzy_t.gzzy001  #語言別
   DEFINE p_os         LIKE type_t.chr1     #系統環境 L/W/M 

   #150901-00021#1 mark(s)
#   DEFINE ls_mdir      STRING
#   DEFINE ls_4rpdir    STRING
#   DEFINE l_sys        STRING                
#   DEFINE l_sql        STRING                    
#   DEFINE l_cnt        LIKE type_t.num5   
#   DEFINE l_module     LIKE gzde_t.gzde002  #模組代碼
   #150901-00021#1 mark(e)

   #150901-00021#1 mark(s)
#   LET ls_4rpdir = NULL
#   IF cl_null(p_rep_code) THEN
#      SELECT gzde002 INTO l_module FROM gzde_t WHERE gzde001 = g_rep_code
#   ELSE
#      SELECT gzde002 INTO l_module FROM gzde_t WHERE gzde001 = p_rep_code
#   END IF 
#   IF NOT cl_null(l_module) THEN
#      LET l_sys = l_module
#      LET l_sys = l_sys.toLowerCase()
#      IF (p_gzgd003 = "Y" OR p_gzgd003 = "c") THEN
#         IF l_sys.getCharAt(1) = "a" THEN
#            LET l_sys = 'c',l_sys.subString(2,l_sys.getLength()) CLIPPED
#            #LET ls_mdir = l_sys
#         ELSE
#            #LET ls_mdir = l_sys
#         END IF
##      ELSE
##         IF l_sys.getCharAt(1) = "c" THEN
##             LET l_sys = l_sys.subString(2,l_sys.getLength())
##             LET l_sql = "SELECT count(gao01) FROM ",s_dbstring("ds") CLIPPED,"gao_file WHERE gao01=?"
##             PREPARE p_replang_pre1 FROM l_sql
##             EXECUTE p_replang_pre1 USING l_zz011 INTO l_cnt
##             FREE p_replang_pre1
##             IF l_cnt=0 THEN 
##                 LET l_sys='a',l_sys
##             ELSE 
##                 LET ls_mdir = l_sys
##             END IF 
##         ELSE 
##            LET ls_mdir = l_sys                
##         END IF 
#      END IF
#      #AP主機LINUX路徑，例:/u1/topprd/erp/axm/4rp/zh_TW
#      IF p_os= "L" THEN
#         LET ls_4rpdir = os.Path.join(os.Path.join(FGL_GETENV(UPSHIFT(l_sys)) CLIPPED,"4rp"),p_lang)
#      END IF
#      #報表主機WINDOWS路徑，例:D:/T100_gr/topprd/axm/4rp/zh_TW
#      IF  p_os="W" THEN
#         LET ls_4rpdir = os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("PATH4RP") CLIPPED,DOWNSHIFT(l_sys) CLIPPED),"4rp"),p_lang)
#      END IF
#      #AP主機MOUNT至報表主機路徑，例:/T100_gr/topprd/axm/4rp/zh_TW      
#      IF p_os= "M" THEN
#         LET ls_4rpdir = os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("MNT4RP") CLIPPED,DOWNSHIFT(l_sys) CLIPPED),"4rp"),p_lang)
#      END IF          
#   END IF
#   
#   RETURN ls_4rpdir
   #150901-00021#1 mark(e)

   
   RETURN cl_rpt_get_4rpdir(p_rep_code,p_gzgd003,p_lang,p_os)   #150901-00021#1 add
END FUNCTION

################################################################################
# Descriptions...: 根據欄位型態維護作業(azzi090)取得報表數值格式
# Memo...........:
# Usage..........: CALL cl_gr_get_numfmt_str() 
#                  RETURNING l_gzgh043
# Input parameter: 
#                : 
# Return code....: l_gzgh043   數值屬性的小數位數
#                : 
# Date & Author..: 2014 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_get_numfmt_str
################################################################################
PUBLIC FUNCTION cl_gr_get_numfmt_str()
   #150901-00021#1 mark(s)
#   DEFINE l_gztd001    DYNAMIC ARRAY OF LIKE gztd_t.gztd001 #欄位屬性編號
#   DEFINE l_num_fmt    DYNAMIC ARRAY OF LIKE gztd_t.gztd001
#   DEFINE l_gztd014    DYNAMIC ARRAY OF LIKE gztd_t.gztd014 #憑證報表欄位格式
#   DEFINE l_gztd015    DYNAMIC ARRAY OF LIKE gztd_t.gztd015 #查詢報表欄位格式
#   DEFINE l_gztd012    DYNAMIC ARRAY OF LIKE gztd_t.gztd012 #報表數值格式
#
#   DEFINE l_sql        STRING
#   DEFINE l_str        STRING
#   DEFINE l_str_int    STRING
#   DEFINE l_str_dec    STRING
#   DEFINE l_substr     STRING
#   DEFINE l_cnt        LIKE type_t.num5
#   DEFINE l_numfmt_cnt LIKE type_t.num5
#   DEFINE l_n          LIKE type_t.num5
#   DEFINE l_int        LIKE type_t.num10    #小數點前位數
#   DEFINE l_dec        LIKE type_t.num10    #小數點後位數
#   DEFINE l_ooef013    LIKE ooef_t.ooef013 
#   DEFINE l_numfmt     STRING
#   DEFINE l_numfmt_x   LIKE gzgg_t.gzgg008   #XG 數值格式
#   DEFINE l_gzgh043    STRING   
   #150901-00021#1 mark(e)
   
   #150901-00021#1 mark(s)
#   LET l_gzgh043 = ""
#   
#   LET l_num_fmt[1] = "N001"
#   LET l_num_fmt[2] = "N003"
#   LET l_num_fmt[3] = "N004"
#   LET l_num_fmt[4] = "N101"
#   LET l_num_fmt[5] = "N102"
#   LET l_num_fmt[6] = "N103"
#   LET l_num_fmt[7] = "N201"
#   LET l_num_fmt[8] = "N202"
#   LET l_num_fmt[9] = "N203"
#   LET l_num_fmt[10] = "N204"
#   LET l_num_fmt[11] = "N205"
#   LET l_num_fmt[12] = "N301"
#   LET l_num_fmt[13] = "N302"
#   LET l_num_fmt[14] = "N303"
#   LET l_num_fmt[15] = "N304"
#   LET l_num_fmt[16] = "N305"
#   LET l_num_fmt[17] = "N401"
#   LET l_num_fmt[18] = "N402"
#   LET l_num_fmt[19] = "N403"
#   LET l_num_fmt[20] = "N601"   
#  
#   LET g_grNumFmt.default = "--,---,---,---,---,---,--&.&&"
#   LET g_grNumFmt.seq = "-------------------&"
#
#   FOR l_cnt = 1 TO l_num_fmt.getlength() 
#      SELECT gztd014,gztd015 INTO l_gztd014[l_cnt],l_gztd015[l_cnt] FROM gztd_t WHERE gztd001 = l_num_fmt[l_cnt]
#      LET l_str_int = l_gztd014[l_cnt] CLIPPED
#      IF cl_null(l_str_int) THEN
#         LET l_str_int = g_grNumFmt.default
#      END IF      
##      CALL cl_gr_get_numfmt(l_num_fmt[l_cnt],"",l_str_int) RETURNING l_numfmt,l_numfmt_x   #150901-00021#1 mark
#      CALL cl_rpt_get_numfmt(l_num_fmt[l_cnt],"",l_str_int) RETURNING l_numfmt,l_numfmt_x   #150901-00021#1 add 
#      LET l_numfmt_cnt = l_numfmt_cnt + 1   
#
#      LET l_gzgh043 = l_gzgh043,"|",l_gztd015[l_cnt] CLIPPED 
#   END FOR
#   LET l_gzgh043 = l_gzgh043.subString(2,l_gzgh043.getLength())
#   
#   LET l_substr = "{0:",g_grNumFmt.default CLIPPED,"}|","{0:",g_grNumFmt.seq CLIPPED,"}"
#   LET l_substr = cl_replace_str(l_substr,"-","#")
#   LET l_substr = cl_replace_str(l_substr,"&","0")
#   LET l_gzgh043 = l_substr,"|",l_gzgh043   
#   
#   RETURN l_gzgh043
   #150901-00021#1 mark(e)
   
   RETURN cl_rpt_get_numfmt_str()
END FUNCTION

################################################################################
# Descriptions...: 取出數值格式
# Memo...........:
# Usage..........: CALL cl_gr_get_numfmt(ps_type,ps_flag,ps_fmt)
#                  RETURNING l_numfmt,l_numfmt_x 
# Input parameter: ps_type   數值格式(例如:N004)
#                : ps_flag   flag(GR傳入G, XG傳入X, LIB傳空值)
#                : ps_fmt    數值格式字串(例如:###.###.##&,&&)
# Return code....: l_numfmt     非XG 數值格式 
#                : l_numfmt_x   XG 數值格式
# Date & Author..: 2014 By 04010
# Modify.........: #150901-00021#1 改用cl_rpt_get_numfmt
################################################################################
PUBLIC FUNCTION cl_gr_get_numfmt(ps_type,ps_flag,ps_fmt)
   DEFINE ps_type         LIKE gztd_t.gztd001
   DEFINE ps_flag         LIKE type_t.chr1
   DEFINE ps_fmt          STRING   #GR的global數值
   DEFINE l_num_fmt       DYNAMIC ARRAY OF LIKE gztd_t.gztd001
   DEFINE l_numfmt        STRING
   DEFINE l_numfmt_x      LIKE gzgg_t.gzgg008  #XG回傳的fmt只能是chr1

   #150901-00021#1 mark(s)
#   CASE ps_type 
#      WHEN 'N001'   #一般小整數
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N001"
#            WHEN "X"
#               LET l_numfmt_x = "N001"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N001 = ps_fmt 
#               END IF
#         END CASE
#      WHEN 'N003'   #一般數值
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N003"
#            WHEN "X"
#               LET l_numfmt_x = "N003"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N003 = ps_fmt 
#               END IF
#         END CASE
#      WHEN 'N004'   #項次
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N004"
#            WHEN "X"
#               LET l_numfmt_x = "N004"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N004 = ps_fmt
#               END IF
#         END CASE
#      WHEN 'N101'   #數量
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N101"
#            WHEN "X"
#               LET l_numfmt_x = "N101"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N101 = ps_fmt
#               END IF
#         END CASE
#      WHEN 'N102'   #重量
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N102"
#            WHEN "X"
#               LET l_numfmt_x = "N102"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N102 = ps_fmt
#               END IF 
#         END CASE
#      WHEN 'N103'   #單位用量/底數/單耗
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N103"
#            WHEN "X"
#               LET l_numfmt_x = "N103"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N103 = ps_fmt
#               END IF 
#         END CASE
#      WHEN 'N201'   #單價
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N201"
#            WHEN "X"
#               LET l_numfmt_x = "N201"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N201 = ps_fmt
#               END IF
#         END CASE
#      WHEN 'N202'   #金額
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N202"
#            WHEN "X" 
#               LET l_numfmt_x = "N202"
#            OTHERWISE
#               LET l_numfmt = ""   
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N202 = ps_fmt
#               END IF 
#         END CASE
#      WHEN 'N203'   #匯率
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N203"
#            WHEN "X"
#               LET l_numfmt_x = "N203"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N203 = ps_fmt
#               END IF
#         END CASE
#      WHEN 'N204'   #成本
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N204"
#            WHEN "X"
#               LET l_numfmt_x = "N204"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N204 = ps_fmt
#               END IF
#         END CASE
#      WHEN 'N205'   #成本金額
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N205"
#            WHEN "X"
#               LET l_numfmt_x = "N205"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N205 = ps_fmt
#               END IF
#         END CASE      
#      WHEN 'N301'   #轉換率/百分比/比率
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N301"
#            WHEN "X"
#               LET l_numfmt_x = "N301"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N301 = ps_fmt
#               END IF
#         END CASE
#      WHEN 'N302'   #稅率
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N302"
#            WHEN "X"
#               LET l_numfmt_x = "N302"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N302 = ps_fmt
#               END IF
#         END CASE
#      WHEN 'N303'   #因數
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N303"
#            WHEN "X"
#               LET l_numfmt_x = "N303"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N303 = ps_fmt
#               END IF
#         END CASE    
#      WHEN 'N304'   #允收水準
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N304"
#            WHEN "X"
#               LET l_numfmt_x = "N304"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N304 = ps_fmt
#               END IF
#         END CASE          
#      WHEN 'N305'   #利率
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N305"
#            WHEN "X"
#               LET l_numfmt_x = "N305"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N305 = ps_fmt
#               END IF
#         END CASE          
#      WHEN 'N401'   #年度
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N401"
#            WHEN "X"
#               LET l_numfmt_x = "N401"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N401 = ps_fmt
#               END IF
#         END CASE           
#      WHEN 'N402'   #期別
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N402"
#            WHEN "X"
#               LET l_numfmt_x = "N402"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N402 = ps_fmt
#               END IF
#         END CASE           
#      WHEN 'N403'   #時數/分數
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N403"
#            WHEN "X"
#               LET l_numfmt_x = "N403"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N403 = ps_fmt
#               END IF
#         END CASE   
#      WHEN 'N601'   #積分
#         CASE ps_flag
#            WHEN "G"
#               LET l_numfmt = "g_grNumFmt.N601"
#            WHEN "X"
#               LET l_numfmt_x = "N601"
#            OTHERWISE
#               LET l_numfmt = ""
#               IF NOT cl_null(ps_fmt) THEN
#                  LET g_grNumFmt.N601 = ps_fmt
#               END IF
#         END CASE           
#      END CASE  
   #150901-00021#1 mark(e)   

   CALL cl_rpt_get_numfmt(ps_type,ps_flag,ps_fmt) RETURNING l_numfmt,l_numfmt_x   #150901-00021#1 add
   RETURN l_numfmt,l_numfmt_x
END FUNCTION

################################################################################
# Descriptions...: 取得報表類型(規格類別) GR/XG
# Memo...........:
# Usage..........: CALL cl_gr_rep_type() 
#                  RETURNING l_gzde003
# Input parameter: 
#                : 
# Return code....: l_gzde003   規格類別
#                : 
# Date & Author..: 2014 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_rep_type
################################################################################
PUBLIC FUNCTION cl_gr_rep_type()
   #DEFINE l_gzde003       LIKE gzde_t.gzde003   #141017 mark
   DEFINE l_gzde005       LIKE gzde_t.gzde005   #141017 add

   #150901-00021#1 mark(s)
#   #SELECT gzde003 INTO l_gzde003 FROM gzde_t   #141017 mark
#   SELECT gzde005 INTO l_gzde005 FROM gzde_t   #141017 add
#    WHERE gzdestus = "Y" AND gzde001 = g_rep_code
#
#   #RETURN l_gzde003   #141017 mark
#   RETURN l_gzde005   #141017 add
   #150901-00021#1 mark(e)
   
   RETURN cl_rpt_rep_type()   #150901-00021#1 add
END FUNCTION

################################################################################
# Descriptions...: GR列印視窗單身填充
# Memo...........:
# Usage..........: CALL cl_gr_b_fill()
#                  
# Input parameter: none
# Return code....: void
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_b_fill()
   DEFINE li_cnt    LIKE type_t.num10
   DEFINE li_cnt_1  LIKE type_t.num5
   DEFINE l_sql     STRING
   DEFINE l_sys     STRING
   DEFINE l_roles   STRING
   DEFINE l_vn_sql  STRING   #180123-00036#1 add

   LET l_roles = cl_replace_str(g_roles,",","','")
   LET l_roles = "('",l_roles,"')"
   CALL g_gzgd.clear()
   LET li_cnt = 1
   LET l_sql = ""
   LET l_sql = "SELECT 'N',gzgdl002,gzgd003,gzgd007,gzgd002,gzgd000,", 
               "       gzgd005,gzgd004,gzgd008,gzgd009,gzgd010,0 ", 
               " FROM gzgd_t left join gzgdl_t on gzgd000 = gzgdl000 AND gzgdl001='", g_lang, "'",
               " WHERE gzgd001='",g_rep_code CLIPPED,"'",
               " AND gzgd007 NOT LIKE '%_sub%'",
               " AND gzgdstus = 'Y' ",
               " AND ((gzgd004='default' AND gzgd005='default')",
               " OR   (gzgd004 IN ", l_roles CLIPPED," AND gzgd005='default')", #顯示符合角色的樣板
               " OR   (gzgd004='default' AND gzgd005='",g_user CLIPPED,"'))"  #顯示符合用戶的樣板
                
   IF NOT cl_null(g_template) THEN
      LET l_sql = l_sql," AND gzgd002='",g_template CLIPPED,"'"
   END IF

   #180213-00002#1 add(s)
   IF NOT cl_null(g_bg4rp) AND g_gzgd002_flag = "Y" THEN
      LET l_sql = l_sql," AND ",g_bg4rp
   END IF
   #180213-00002#1 add(e)

#   SELECT COUNT(*) INTO li_cnt_1 FROM gzgd_t WHERE gzgdstus = 'Y' AND gzgd001 = g_rep_code AND gzgd007 NOT LIKE '%_sub%' AND (gzgd003 = "Y" OR gzgd003 = "c")   #170525-00105#1 mark
   SELECT COUNT(1) INTO li_cnt_1 FROM gzgd_t WHERE gzgdstus = 'Y' AND gzgd001 = g_rep_code AND gzgd007 NOT LIKE '%_sub%' AND (gzgd003 = "Y" OR gzgd003 = "c")    #170525-00105#1 add
   IF li_cnt_1 > 0 THEN
      LET l_sql = l_sql," AND (gzgd003 = 'Y' OR gzgd003 = 'c') "
   ELSE
      LET l_sql = l_sql," AND (gzgd003 = 'N' OR gzgd003 = 's') "
   END IF

   LET l_vn_sql = l_sql, " AND gzgd012 <> 'N' "   #180123-00036#1 add

   #若語言別為越文，則一併在報表樣板清單中顯示中越文樣板，否則不顯示
   IF g_lang <> "vi_VN" THEN
      LET l_sql = l_sql," AND gzgd012 = 'N' "
   END IF   

   LET l_sql = l_sql," ORDER BY gzgd003 DESC,gzgd007" 
   DECLARE cl_gzgd_cs1 CURSOR FROM l_sql

   FOREACH cl_gzgd_cs1 INTO g_gzgd[li_cnt].*
      IF SQLCA.sqlcode THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = SQLCA.sqlcode
         LET g_errparam.extend = 'FOREACH:'
         LET g_errparam.popup = TRUE
         CALL cl_err()

         EXIT FOREACH
      END IF
      IF cl_null(g_gzgd[li_cnt].gzgd008) THEN LET g_gzgd[li_cnt].gzgd008 = 'N' END IF
      IF li_cnt = 1 THEN LET g_gzgd[li_cnt].pick ='Y' END  IF  
         LET li_cnt = li_cnt + 1
   END FOREACH

   CLOSE cl_gzgd_cs1
   FREE cl_gzgd_cs1
   CALL g_gzgd.deleteElement(li_cnt)
   LET g_rec_b = li_cnt - 1
   
   #180123-00036#1 add(s)
   IF g_rec_b = 0 THEN
      #非越文語系登入, 只有中越樣板需提示
      IF g_lang <> 'vi_VN' THEN
         LET l_vn_sql = "SELECT COUNT(1) FROM (",l_vn_sql,") tsed"
         PREPARE cl_gr_getvn_cnt_pre FROM l_vn_sql
         EXECUTE cl_gr_getvn_cnt_pre INTO li_cnt_1
         FREE cl_gr_getvn_cnt_pre
         IF li_cnt_1 > 0 THEN
            LET g_rep_flag = TRUE

            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = 'lib-00441'
            LET g_errparam.extend = g_template
            LET g_errparam.popup = TRUE
            CALL cl_err()
         END IF
      END IF
   END IF
   #180123-00036#1 add(e)
   
   #依優先序重新排序陣列內容
   #IF g_rec_b > 1 THEN
      CALL cl_gr_b_fill_sort()
   #END IF 
END FUNCTION

################################################################################
# Descriptions...: 產生GR的Sample Data
# Memo...........:
# Usage..........: LET handler = cl_gr_sampledata(ls_dir,g_gzgd[l_ac].gzgd007)
#                  
# Input parameter: p_dir          存放目錄路徑
#                : p_gzgd007      4rp檔名
# Return code....: om.SaxDocumentHandler    SAX文件處理器物件
#                : 
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_sampledata(p_dir,p_gzgd007)
   DEFINE handler         om.SaxDocumentHandler
   DEFINE p_gzgd007       LIKE gzgd_t.gzgd007
   DEFINE p_dir           STRING
   DEFINE l_dir           STRING
   DEFINE ls_filename     STRING
   DEFINE ls_message      STRING

   #移除目錄語言別
   LET p_dir = cl_str_replace(p_dir,g_rlang,"")        #160608-00012#1 add
   LET p_dir = p_dir.subString(1,p_dir.getLength())    #160608-00012#1 add
   #create dir
   LET l_dir = os.Path.join( p_dir, "sampledata")
   IF NOT os.Path.exists(l_dir) THEN     #預設一定有這個資料夾
      IF os.Path.mkdir(l_dir) THEN
         IF os.Path.chrwx(l_dir ,511) THEN
	    #開放權限
         END IF
      ELSE
         DISPLAY "Create folder ",l_dir," fails."
         RETURN NULL
      END IF
   END IF

   -- make 4rp
   LET ls_filename = cl_gr_make4rp(l_dir, p_gzgd007)

   -- load the 4rp file
   IF NOT fgl_report_loadCurrentSettings(ls_filename) THEN
      DISPLAY "Load report error"
   ELSE
      CALL cl_gr_del4rp(ls_filename)
   END IF

   LET ls_filename = p_gzgd007 , ".sampledata"
   LET ls_filename = os.Path.join(l_dir ,ls_filename)
   LET g_sampledata_path = ls_filename                  #160608-00012#1 add

   -- write to 4rp with empty data
   CALL fgl_report_selectDevice("SVG")
   CALL fgl_report_selectPreview(0)

   -- Generate an XML data file:
   #170725-00086#1 mark(s)
#   IF g_gr_progress = 'Y' THEN
#       LET ls_message = "processing sampledata...."
#       CALL cl_progress_ing(ls_message)
#   END IF
   #170725-00086#1 mark(e)
   LET handler = fgl_report_createProcessLevelDataFile(ls_filename)
   #170725-00086#1 mark(s)   
#   IF g_gr_progress = 'Y' THEN
#       LET ls_message = "finish sampledata...."   
#       CALL cl_progress_ing(ls_message)   
#       #CALL cl_progressing("finish sampledata...." )
#   END IF
   #170725-00086#1 mark(e)   
   IF os.Path.chrwx(ls_filename ,511) THEN
      #開放權限
   END IF
   RETURN handler
END FUNCTION

################################################################################
# Descriptions...: 刪除4rp樣板檔
# Memo...........:
# Usage..........: CALL cl_gr_del4rp(ls_filename)
#                  
# Input parameter: ls_filename    檔案路徑
# Return code....: void
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_del4rp(ls_filename)
   DEFINE ls_filename     STRING

   IF os.Path.exists(ls_filename) THEN
      IF NOT os.Path.delete(ls_filename) THEN
	      DISPLAY "Cannot delete" , ls_filename
      END IF
   END IF
END FUNCTION

################################################################################
# Descriptions...: 產生4rp樣板檔
# Memo...........:
# Usage..........: LET ls_filename = cl_gr_make4rp(l_dir, p_gzgd007)
#                  
# Input parameter: p_dir          存放目錄路徑
#                : p_gzgd007      4rp檔名
# Return code....: STRING         4rp檔路徑
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_make4rp(p_dir,p_gzgd007)
   DEFINE p_dir           STRING
   DEFINE p_gzgd007       LIKE gzgd_t.gzgd007
   DEFINE lc_channel      base.Channel
   DEFINE ls_filename     STRING
   DEFINE ls_rdd          STRING
   DEFINE ls_report_name  STRING
   DEFINE ls_temp         STRING


   -- build 4rp xml
   LET ls_filename = os.Path.join(p_dir, p_gzgd007 || '.4rp')
   CALL cl_gr_del4rp(ls_filename)

   LET ls_rdd = p_gzgd007 , ".rdd"
   LET ls_report_name = p_gzgd007 , "_rep"

   LET lc_channel = base.Channel.create()
   CALL lc_channel.openFile( ls_filename CLIPPED, "a" )
   CALL lc_channel.setDelimiter("")
   CALL lc_channel.write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
   CALL lc_channel.write("<report:Report xmlns:rtl=\"http://www.4js.com/2004/RTL\" xmlns:report=\"http://www.4js.com/2007/REPORT\" xmlns=\"http://www.4js.com/2004/PXML\" gstVersion=\"2320f\" version=\"5.00\"> ")
   CALL lc_channel.write("<report:Settings RWPageWidth=\"a4width\" RWPageLength=\"a4length\" RWLeftMargin=\"1.3cm\" RWTopMargin=\"1.3cm\" RWRightMargin=\"1.3cm\" RWBottomMargin=\"1.3cm\"> ")
   CALL lc_channel.write("     <report:FormatList>       ")
   CALL lc_channel.write("        <report:Format-SVG/>   ")
   CALL lc_channel.write("        <report:Format-PDF/>   ")
   CALL lc_channel.write("        <report:Format-image/> ")
   CALL lc_channel.write("     </report:FormatList>      ")
   CALL lc_channel.write("</report:Settings>      ")
   LET ls_temp = "   <report:Data RWDataLocation=\"", ls_rdd ,
                         "\" RWFglReportName=\"" ,ls_report_name, "\"/>"
   CALL lc_channel.write(ls_temp)
   CALL lc_channel.write("     <rtl:stylesheet> ")
   CALL lc_channel.write("         <PXML>       ")
   CALL lc_channel.write("            <rtl:match nameConstraint=\"Report\"> ")
   CALL lc_channel.write("               <MINIPAGE name=\"Page Root\" width=\"max\" length=\"max\"> ")
   CALL lc_channel.write("                    <rtl:match nameConstraint=\"Group\" minOccurs=\"0\" maxOccurs=\"unbounded\"> ")
   CALL lc_channel.write("                       <rtl:match nameConstraint=\"OnEveryRow\" minOccurs=\"0\" maxOccurs=\"unbounded\"/> ")
   CALL lc_channel.write("                    </rtl:match> ")
   CALL lc_channel.write("               </MINIPAGE> ")
   CALL lc_channel.write("            </rtl:match> ")
   CALL lc_channel.write("         </PXML>      ")
   CALL lc_channel.write("     </rtl:stylesheet> ")
   CALL lc_channel.write("</report:Report> ")
   CALL lc_channel.close()
   RETURN ls_filename

END FUNCTION

################################################################################
# Descriptions...: 關閉GR報表詢問視窗
# Memo...........:
# Usage..........: CALL cl_gr_close_report()
#                  
# Input parameter: void
# Return code....: none
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_close_report()
   DEFINE l_rootnode	      om.DomNode
   DEFINE l_nodes	         om.NodeList
   DEFINE l_chmodname      STRING   
   DEFINE l_pdfname        LIKE type_t.chr300  #歷史報表記錄檔-歷史報表檔名
   DEFINE l_sb             base.StringBuffer   
   DEFINE l_oojp004        LIKE oojp_t.oojp004
   DEFINE l_roles_arr      DYNAMIC ARRAY OF STRING
   DEFINE l_roles          STRING
   DEFINE l_tok_param      base.StringTokenizer
   DEFINE i                LIKE type_t.num5
   DEFINE l_oojp003        LIKE oojp_t.oojp003

#   DISPLAY "g_savefilename:",g_savefilename

   IF NOT cl_null(g_savefilename) THEN   #170406-00030#1 add
     LET l_pdfname=""
     LET l_pdfname=os.Path.basename(g_savefilename)
    
#     LET l_chmodname =os.Path.join(FGL_GETENV("TEMPDIR"),os.Path.basename(g_savefilename))
#     LET l_chmodname=os.Path.join(os.Path.join(FGL_GETENV("MNT4RP"),"history"),os.Path.basename(g_savefilename))             #150916-00002#1 mark
#     LET l_chmodname=os.Path.join(os.Path.join(cl_rpt_get_env_global("MNT4RP"),"history"),os.Path.basename(g_savefilename))   #150916-00002#1 add   #170725-00086#1 mark
     LET l_chmodname=os.Path.join(os.Path.join(cl_rpt_get_env_global("PATH4RP"),"history"),os.Path.basename(g_savefilename))   #150916-00002#1 add   #170725-00086#1 add
#     IF os.Path.exists(l_chmodname) THEN                                     #170725-00086#1 mark
#        IF os.Path.chrwx(l_chmodname,511) THEN END IF  #修改歷史留檔pdf權限   #170725-00086#1 mark
        IF l_chmodname.getIndexOf("pdf",1) <>0 THEN
           #寫入歷史報表歷程檔
           LET l_oojp004 = cl_get_current()
    
           #取出職能角色，若有admin-01，則取admin-01，若無就取第一個
           LET l_tok_param = base.StringTokenizer.createExt(g_roles,",","",TRUE)
           CALL l_roles_arr.clear()
           LET i = 0
           WHILE l_tok_param.hasMoreTokens()
              LET i = i + 1
              LET l_roles_arr[i] = l_tok_param.nextToken()
           END WHILE
          
           LET l_roles = ""
           FOR i = 1 TO l_roles_arr.getLength()
              IF l_roles_arr[i] = "admin-01" THEN
                 LET l_roles = l_roles_arr[i]
              END IF
           END FOR
           
           IF l_roles IS NULL THEN
              LET l_roles = l_roles_arr[1]
           END IF
    
           LET l_oojp003 = l_roles
    
           #歷史報表記錄檔
           INSERT INTO oojp_t(oojpent,oojpsite,oojp001,oojp002,oojp003,oojp004,oojp005,oojp006)    
           VALUES (g_enterprise,g_site,g_prog,g_user,l_oojp003,l_oojp004,l_pdfname,'Y')
           
           IF SQLCA.sqlcode THEN
              INITIALIZE g_errparam TO NULL 
              LET g_errparam.extend = "oojp_t" 
              LET g_errparam.code   = SQLCA.sqlcode 
              LET g_errparam.popup  = FALSE 
              CALL cl_err()
              CALL s_transaction_end('N','0')
           END IF
        END IF 
     END IF 
#   END IF   #170406-00030#1 add   #170725-00086#1 mark
 
   LET g_savefilename = ""
   
   LET l_rootnode = ui.Interface.getRootNode()
   LET l_nodes = l_rootnode.selectByPath("//Window[@name=\"cl_gr_w\"]")
   IF l_nodes.getLength() >= 1 THEN
      CLOSE WINDOW cl_gr_w
   END IF
END FUNCTION

################################################################################
# Descriptions...: 開啟GR mail視窗
# Memo...........:
# Usage..........: CALL cl_gr_mail()
#                  
# Input parameter: void
# Return code....: none
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_mail()
   #190408-00033#2 mark(s)
#   DEFINE   lr_mail        DYNAMIC ARRAY OF RECORD
#              subject        LIKE gzpe_t.gzpe005,
#              context        LIKE gzpe_t.gzpe006,
#              #javamail預設收件人         
#              oojn005        LIKE oojn_t.oojn005, #1:收件者 2:副本 3:密件副本
#              oojn011        LIKE oojn_t.oojn011, #連絡對象類型   #150803-00008#3 add              
#              oojn003        LIKE oojn_t.oojn003, #員工編號   #150611-00032#3 add              
#              oojn004        LIKE oojn_t.oojn004, #姓名          
#              oojn006        LIKE oojn_t.oojn006, #Mail address
#              oojn012        LIKE oojn_t.oojn012, #進階設定   #190107-00075#6 add
#              oojn007        LIKE oojn_t.oojn007, #報表mail檔案格式
#              gzxa001        LIKE gzxa_t.gzxa001  #使用者編號 #150922-00009#1 add
#                           END RECORD
   #190408-00033#2 mark(e)
   #190408-00033#2 add(s)
   DEFINE lr_mail        DYNAMIC ARRAY OF RECORD
              subject        LIKE gzpe_t.gzpe005,
              context        LIKE gzpe_t.gzpe006,
              #javamail預設收件人         
              oojn005        LIKE gzpe_t.gzpe007,  #1:收件者 2:副本 3:密件副本
              oojn011        LIKE gzpe_t.gzpe012,  #聯絡對象類型
              oojn003        LIKE gzpe_t.gzpe004,  #收件人員工編號 
              oojn004        LIKE gzpe_t.gzpe008,  #姓名 
              oojn006        LIKE gzpe_t.gzpe009,  #Mail address
              oojn012        LIKE gzpe_t.gzpe013,  #進階設定
              oojn007        LIKE gzpe_t.gzpe010,  #報表mail檔案格式   
              gzxa001        LIKE gzxa_t.gzxa001  #使用者編號
                           END RECORD   
   #190408-00033#2 add(e)
   
                           
                           
   DEFINE   lst_maillist   base.StringTokenizer
   DEFINE   ls_mail        STRING
   DEFINE   li_cnt         LIKE type_t.num5
   DEFINE   li_i           LIKE type_t.num5
   DEFINE   li_idx         LIKE type_t.num5
   DEFINE   l_sql          STRING
   DEFINE   l_subject      VARCHAR(300)          #150709-00023#1 add
   DEFINE   l_context      STRING                #150709-00023#1 add
   #190107-00075#6 add(s)
   DEFINE   l_maillist     RECORD                #進階設定JSON
              recipient    STRING,
              cc           STRING,
              bcc          STRING
                           END RECORD
   #190107-00075#6 add(e)

   CALL lr_mail.clear()   #150709-00023#1 add
   IF g_bgjob = "Y" THEN
      IF g_schedule_d.getLength() > 0 THEN
         DISPLAY "立即於背景執行"  #190606-00010#1 add
         FOR li_idx = 1 TO g_schedule_d.getLength()
            IF g_schedule_d[li_idx].gzpe003 = "Z" OR g_schedule_d[li_idx].gzpe003 = "Y" THEN   #150921-00060#1 add   
               LET lr_mail[li_idx].oojn011 = g_schedule_d[li_idx].gzpe012   #150803-00008#3 add 
               #150709-00023#1 add(s)
               LET lr_mail[li_idx].oojn005 = g_schedule_d[li_idx].gzpe007
               LET lr_mail[li_idx].oojn003 = g_schedule_d[li_idx].gzpe004
               LET lr_mail[li_idx].oojn004 = g_schedule_d[li_idx].gzpe008
               IF NOT cl_null(g_schedule_d[li_idx].gzpe009) THEN
                  LET lr_mail[li_idx].oojn006 = g_schedule_d[li_idx].gzpe009              
               ELSE
#                  LET lr_mail[li_idx].oojn006 = s_aooi350_get_user_contact(g_schedule_d[li_idx].gzpe004,'4')   #150803-00008#3 mark
#                  LET lr_mail[li_idx].oojn006 = s_aooi350_get_cont_ppl_email(g_schedule_d[li_idx].gzpe012,g_schedule_d[li_idx].gzpe004)   #150803-00008#3 add   #161222-00006#1 mark
#                  LET lr_mail[li_idx].oojn006 = cl_s_aooi350_get_cont_ppl_email(g_schedule_d[li_idx].gzpe012,g_schedule_d[li_idx].gzpe004)                       #161222-00006#1 add   #170929-00002#1 mark
                  LET lr_mail[li_idx].oojn006 = cl_gr_get_cont_ppl_email(g_schedule_d[li_idx].gzpe012,g_schedule_d[li_idx].gzpe004)                               #161222-00006#1 add   #170929-00002#1 add
               END IF
               LET lr_mail[li_idx].oojn012 = g_schedule_d[li_idx].gzpe013   #190107-00075#6 add
               LET lr_mail[li_idx].oojn007 = g_schedule_d[li_idx].gzpe010
               #150709-00023#1 add(e)
               #190606-00010#1 add(s)
               LET lr_mail[li_idx].subject =  g_schedule_d[li_idx].gzpe005 
               LET lr_mail[li_idx].context =  g_schedule_d[li_idx].gzpe006
               #190606-00010#1 add(e)               
               #150709-00023#1 mark(s)
#               LET g_schedule_d[li_idx].gzpe009 = s_aooi350_get_user_contact(g_schedule_d[li_idx].gzpe004,'4')   #150611-00032#3 add
#               CASE g_schedule_d[li_idx].gzpe007
#                  WHEN "1"
#                     LET g_sendmail[li_idx].receiver = g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010
#                  WHEN "2"
#                     LET g_sendmail[li_idx].cc = g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010
#                  WHEN "3"
#                     LET g_sendmail[li_idx].bcc = g_schedule_d[li_idx].gzpe009,":",g_schedule_d[li_idx].gzpe008,":",g_schedule_d[li_idx].gzpe010
#               END CASE
#               LET g_sendmail[li_idx].format = g_schedule_d[li_idx].gzpe010
#               LET g_sendmail[li_idx].subject = g_schedule_d[li_idx].gzpe005
#               LET g_sendmail[li_idx].context = g_schedule_d[li_idx].gzpe006
               #150709-00023#1 mark(e)
            END IF #150921-00060#1 add  
         END FOR
         #190606-00010#1 mark(s)             
#         LET l_subject = g_schedule_d[li_idx].gzpe005   #150709-00023#1 add
#         LET l_context = g_schedule_d[li_idx].gzpe006   #150709-00023#1 add
         #190606-00010#1 mark(e)
      ELSE 
#         LET l_sql = " SELECT gzpe005,gzpe006,gzpe007,gzpe008,gzpe009,gzpe010 FROM gzpe_t WHERE gzpeent=? AND gzpe001=? AND gzpe002=? AND gzpe003 = 'Z' "   #150611-00032#3 mark
#         LET l_sql = " SELECT gzpe005,gzpe006,gzpe007,gzpe004,gzpe008,gzpe009,gzpe010 FROM gzpe_t WHERE gzpeent=? AND gzpe001=? AND gzpe002=? AND gzpe003 = 'Z' " #150611-00032#3 add   #150803-00008#3 mark
#         LET l_sql = " SELECT gzpe005,gzpe006,gzpe007,gzpe012,gzpe004,gzpe008,gzpe009,gzpe010 FROM gzpe_t WHERE gzpeent=? AND gzpe001=? AND gzpe002=? AND gzpe003 = 'Z' " #150803-00008#3 add   #190107-00075#6 mark
         DISPLAY "指定時間於背景執行"   #190606-00010#1 add
         #190107-00075#6 add(s)
         LET l_sql = " SELECT gzpe005,gzpe006,gzpe007,gzpe012,gzpe004,",
                     "        gzpe008,gzpe009,gzpe013,gzpe010 ",   #add gzpe013
                     "   FROM gzpe_t WHERE gzpeent=? AND gzpe001=? AND gzpe002=? AND gzpe003 = 'Z' "
         #190107-00075#6 add(e)
         LET l_sql = l_sql," ORDER BY gzpe011 "   #150709-00023#1 add
         DECLARE cl_gr_mail_gzpe_cs CURSOR FROM l_sql
     
         LET li_idx = 1
         FOREACH cl_gr_mail_gzpe_cs USING g_enterprise,g_schedule_id,g_schedule_seq INTO lr_mail[li_idx].*
            IF cl_null(lr_mail[li_idx].oojn006) THEN
#               LET lr_mail[li_idx].oojn006 = s_aooi350_get_user_contact(lr_mail[li_idx].oojn003,'4')   #150803-00008#3 mark
#               LET lr_mail[li_idx].oojn006 = s_aooi350_get_cont_ppl_email(lr_mail[li_idx].oojn011,lr_mail[li_idx].oojn003)   #150803-00008#3 add   #161222-00006#1 mark
#               LET lr_mail[li_idx].oojn006 = cl_s_aooi350_get_cont_ppl_email(lr_mail[li_idx].oojn011,lr_mail[li_idx].oojn003)                       #161222-00006#1 add   #170929-00002#1 mark
               LET lr_mail[li_idx].oojn006 = cl_gr_get_cont_ppl_email(lr_mail[li_idx].oojn011,lr_mail[li_idx].oojn003)                               #161222-00006#1 add   #170929-00002#1 add
            END IF   
            #150709-00023#1 mark(s)
#            CASE lr_mail[li_idx].oojn005
#               WHEN "1"
##                  LET g_sendmail[li_idx].receiver = lr_mail[li_idx].oojn006,":",lr_mail[li_idx].oojn004,":",lr_mail[li_idx].oojn007   
##                  DISPLAY "g_sendmail[li_idx].receiver:",g_sendmail[li_idx].receiver
#               WHEN "2"
#                  LET g_sendmail[li_idx].cc = lr_mail[li_idx].oojn006,":",lr_mail[li_idx].oojn004,":",lr_mail[li_idx].oojn007
##                  DISPLAY "g_sendmail[li_idx].cc:",g_sendmail[li_idx].cc
#               WHEN "3"
#                  LET g_sendmail[li_idx].bcc = lr_mail[li_idx].oojn006,":",lr_mail[li_idx].oojn004,":",lr_mail[li_idx].oojn007
##                  DISPLAY "g_sendmail[li_idx].bcc:",g_sendmail[li_idx].bcc
#            END CASE
#            
#            LET g_sendmail[li_idx].format = lr_mail[li_idx].oojn007
#            LET g_sendmail[li_idx].subject = lr_mail[li_idx].subject
#            LET g_sendmail[li_idx].context = lr_mail[li_idx].context
            #150709-00023#1 mark(e)            
            LET li_idx = li_idx + 1
         END FOREACH
         
         #150709-00023#1 add(s)
         IF NOT cl_null(g_mail_subject) THEN
            LET l_subject = g_mail_subject             
         ELSE
            #LET l_subject = lr_mail[li_idx].subject   #190606-00010#1 mark
            LET l_subject = lr_mail[1].subject   #190606-00010#1 add            
         END IF
         IF NOT cl_null(g_mail_context) THEN
            LET l_context = g_mail_context
         ELSE
            LET l_context = lr_mail[li_idx].context
         END IF             
         #150709-00023#1 add(e)   
         CLOSE cl_gr_mail_gzpe_cs
         FREE cl_gr_mail_gzpe_cs
      END IF
   ELSE
      CALL cl_gr_selmaillist() RETURNING g_receiver_list, g_cc_list, g_bcc_list
                                         ,g_adv_list   #190107-00075#6 add
                                         ,l_subject,l_context     #150709-00023#1 add
#      IF cl_null(g_receiver_list) AND cl_null(g_cc_list) AND cl_null(g_bcc_list) THEN   #190107-00075#6 mark
      IF cl_null(g_receiver_list) AND cl_null(g_cc_list) AND cl_null(g_bcc_list) AND cl_null(g_adv_list) THEN   #190107-00075#6 add
         RETURN
      END IF
   
      #拆解欲送信的GR format, 以便產生多次javamail寄送
      LET li_cnt = 1
      IF NOT cl_null(g_receiver_list) THEN
         LET lst_maillist = base.StringTokenizer.create(g_receiver_list, ";")
         WHILE lst_maillist.hasMoreTokens()
            LET ls_mail = lst_maillist.nextToken()
            IF NOT cl_null(ls_mail) THEN
               LET lr_mail[li_cnt].oojn005 = "1"
               CALL cl_gr_splite(ls_mail) RETURNING lr_mail[li_cnt].oojn006,lr_mail[li_cnt].oojn004,lr_mail[li_cnt].oojn007,lr_mail[li_cnt].gzxa001 #排程 #150922-00009#1 add gzxa001
               LET li_cnt = li_cnt + 1
            END IF
         END WHILE
      END IF
      IF NOT cl_null(g_cc_list) THEN
         LET lst_maillist = base.StringTokenizer.create(g_cc_list, ";")
         WHILE lst_maillist.hasMoreTokens()
            LET ls_mail = lst_maillist.nextToken()
            IF NOT cl_null(ls_mail) THEN
               LET lr_mail[li_cnt].oojn005 = "2"
               CALL cl_gr_splite(ls_mail) RETURNING lr_mail[li_cnt].oojn006,lr_mail[li_cnt].oojn004,lr_mail[li_cnt].oojn007,lr_mail[li_cnt].gzxa001 #排程 #150922-00009#1 add gzxa001
               LET li_cnt = li_cnt + 1
            END IF
         END WHILE
      END IF
      IF NOT cl_null(g_bcc_list) THEN
         LET lst_maillist = base.StringTokenizer.create(g_bcc_list, ";")
         WHILE lst_maillist.hasMoreTokens()
            LET ls_mail = lst_maillist.nextToken()
            IF NOT cl_null(ls_mail) THEN
               LET lr_mail[li_cnt].oojn005 = "3"
               CALL cl_gr_splite(ls_mail) RETURNING lr_mail[li_cnt].oojn006,lr_mail[li_cnt].oojn004,lr_mail[li_cnt].oojn007,lr_mail[li_cnt].gzxa001 #排程 #150922-00009#1 add gzxa001
               LET li_cnt = li_cnt + 1
            END IF
         END WHILE
      END IF
      #190107-00075#6 add(s)
      IF NOT cl_null(g_adv_list) THEN
         LET lst_maillist = base.StringTokenizer.create(g_adv_list, "|")
         WHILE lst_maillist.hasMoreTokens()
            LET ls_mail = lst_maillist.nextToken()
            IF NOT cl_null(ls_mail) THEN
               LET lr_mail[li_cnt].oojn005 = "4"
               #用&作為收件者與附件格式的分隔符號
               LET lr_mail[li_cnt].oojn012 = ls_mail.subString(1,ls_mail.getIndexOf("&",1)-1)
               LET lr_mail[li_cnt].oojn007 = ls_mail.subString(ls_mail.getIndexOf("&",1)+1,ls_mail.getLength())
               LET li_cnt = li_cnt + 1
            END IF
         END WHILE  
      END IF      
      #190107-00075#6 add(e)
      
      CALL lr_mail.deleteElement(li_cnt)
      #150709-00023#1 mark(s)
#      CALL g_sendmail.clear()
#      FOR li_cnt = 1 TO lr_mail.getLength()
#          LET li_idx = 0
#          FOR li_i = 1 TO g_sendmail.getLength()
#              IF lr_mail[li_cnt].oojn007 = g_sendmail[li_i].format THEN
#                 LET li_idx = li_i
#                 EXIT FOR
#              END IF
#          END FOR
#          IF li_idx = 0 THEN
#             LET li_idx = g_sendmail.getLength() + 1
#             LET g_sendmail[li_idx].format = lr_mail[li_cnt].oojn007
#          END IF
#          CASE lr_mail[li_cnt].oojn005
#             WHEN "1"
#                LET g_sendmail[li_idx].receiver = g_sendmail[li_idx].receiver,
#                                                  lr_mail[li_cnt].oojn006 CLIPPED,":", 
#                                                  lr_mail[li_cnt].oojn004 CLIPPED,";"  
#             WHEN "2"
#                LET g_sendmail[li_idx].cc = g_sendmail[li_idx].cc,
#                                            lr_mail[li_cnt].oojn006 CLIPPED,":",  
#                                            lr_mail[li_cnt].oojn004 CLIPPED,";"  
#             WHEN "3"
#                LET g_sendmail[li_idx].bcc = g_sendmail[li_idx].bcc,
#                                            lr_mail[li_cnt].oojn006 CLIPPED,":", 
#                                            lr_mail[li_cnt].oojn004 CLIPPED,";"  
#          END CASE
#      END FOR
      #150709-00023#1 mark(e)
   END IF

   LET l_context = cl_str_replace(l_context,"\n","<br>")   #置換換行符號 #170602-00007#1 add
   
   #150709-00023#1 add(s)
   CALL g_sendmail.clear()
   FOR li_cnt = 1 TO lr_mail.getLength()
       LET li_idx = 0
       #190107-00075#6 mark(s)
       #原GR會合併寄送，為統一規則故mark，設定進階設定時才合併寄送
#       FOR li_i = 1 TO g_sendmail.getLength()
#           IF lr_mail[li_cnt].oojn007 = g_sendmail[li_i].format THEN
#              LET li_idx = li_i
#              EXIT FOR
#           END IF
#       END FOR
       #190107-00075#6 mark(e)
       IF li_idx = 0 THEN
          LET li_idx = g_sendmail.getLength() + 1
          LET g_sendmail[li_idx].format = lr_mail[li_cnt].oojn007
          #190606-00010#1 add(s)
          IF g_bgjob = "Y" THEN
             #背景寄送
             LET g_sendmail[li_idx].subject = lr_mail[li_cnt].subject
             LET g_sendmail[li_idx].context = lr_mail[li_cnt].context
          ELSE
             #開窗寄送
             LET g_sendmail[li_idx].subject = l_subject
             LET g_sendmail[li_idx].context = l_context
          END IF
          #190606-00010#1 add(e)
          #190606-00010#1 mark(s)
#          LET g_sendmail[li_idx].subject = l_subject   #150709-00023#1 add   
#          LET g_sendmail[li_idx].context = l_context   #150709-00023#1 add
          #190606-00010#1 mark(e)
       END IF
       CASE lr_mail[li_cnt].oojn005
          WHEN "1"
             LET g_sendmail[li_idx].receiver = g_sendmail[li_idx].receiver,
                                               lr_mail[li_cnt].oojn006 CLIPPED,":", 
                                               lr_mail[li_cnt].oojn004 CLIPPED,":",
                                               lr_mail[li_cnt].gzxa001 CLIPPED,";"  #150922-00009#1 add
          WHEN "2"
             LET g_sendmail[li_idx].cc = g_sendmail[li_idx].cc,
                                         lr_mail[li_cnt].oojn006 CLIPPED,":",  
                                         lr_mail[li_cnt].oojn004 CLIPPED,":",
                                         lr_mail[li_cnt].gzxa001 CLIPPED,";"        #150922-00009#1 add
          WHEN "3"
             LET g_sendmail[li_idx].bcc = g_sendmail[li_idx].bcc,
                                         lr_mail[li_cnt].oojn006 CLIPPED,":", 
                                         lr_mail[li_cnt].oojn004 CLIPPED,":",
                                         lr_mail[li_cnt].gzxa001 CLIPPED,";"        #150922-00009#1 add
          #190107-00075#6 add(s)
          WHEN "4"
             CALL util.JSON.parse(lr_mail[li_cnt].oojn012,l_maillist)
             #收件人
             IF NOT cl_null(l_maillist.recipient) THEN
                LET g_sendmail[li_idx].receiver = g_sendmail[li_idx].receiver,
                                                  l_maillist.recipient CLIPPED #,":", 
                                                  #"",":",
                                                  #lr_mail[li_cnt].gzxa001 CLIPPED,";"
                                                  ,";"
             END IF
             #副本
             IF NOT cl_null(l_maillist.cc) THEN
                LET g_sendmail[li_idx].cc = g_sendmail[li_idx].cc,
                                            l_maillist.cc CLIPPED #,":", 
                                            #"",":",
                                            #lr_mail[li_cnt].gzxa001 CLIPPED,";"
                                            ,";"
             END IF
             #密件副本
             IF NOT cl_null(l_maillist.bcc) THEN
                LET g_sendmail[li_idx].bcc = g_sendmail[li_idx].bcc,
                                             l_maillist.bcc CLIPPED #,":", 
                                             #"",":",
                                             #lr_mail[li_cnt].gzxa001 CLIPPED,";"
                                             ,";"
             END IF
          #190107-00075#6 add(e)
       END CASE
   END FOR
   
   LET g_receiver_list = NULL   #第二次開窗應清空
   LET g_cc_list = NULL         #第二次開窗應清空  
   LET g_bcc_list = NULL        #第二次開窗應清空
   #150709-00023#1 add(e)
   LET g_adv_list = NULL   #190107-00075#6 add
      
   IF g_sendmail.getLength() > 0 THEN
      LET g_mail_flag = TRUE
      LET g_mail_cnt = 0   
    END IF
END FUNCTION

################################################################################
# Descriptions...: 設定GR mail收件者、副本、密件副本清單
# Memo...........:
# Usage..........: CALL cl_gr_selmaillist() 
#                  RETURNING g_receiver_list, g_cc_list, g_bcc_list
# Input parameter: void
# Return code....: g_receiver_list  收件者清單
#                : g_cc_list        副本清單 
#                : g_bcc_list       密件副本清單        
# Date & Author..: 2014/10/1 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_selmaillist()
   #190408-00033#2 mark(s)
#   DEFINE   lr_mail        DYNAMIC ARRAY OF RECORD   
#              oojn005      LIKE oojn_t.oojn005,  #1:收件者 2:副本 3:密件副本
#              oojn011      LIKE oojn_t.oojn011,  #聯絡對象類型     #150803-00008#3 add                 
#              oojn003      LIKE oojn_t.oojn003,  #收件人員工編號   #150803-00008#3 add
#              oojn004      LIKE oojn_t.oojn004,  #姓名 
#              oojn006      LIKE oojn_t.oojn006,  #Mail address
#              oojn012      LIKE oojn_t.oojn012,  #進階設定   #190107-00075#6 add
#              oojn007      LIKE oojn_t.oojn007,  #報表mail檔案格式               
#              ready        LIKE type_t.chr1
#                           END RECORD
   #190408-00033#2 mark(e)
   #190408-00033#2 add(s)
   DEFINE lr_mail        DYNAMIC ARRAY OF RECORD
              oojn005      LIKE gzpe_t.gzpe007,  #1:收件者 2:副本 3:密件副本
              oojn011      LIKE gzpe_t.gzpe012,  #聯絡對象類型
              oojn003      LIKE gzpe_t.gzpe004,  #收件人員工編號 
              oojn004      LIKE gzpe_t.gzpe008,  #姓名 
              oojn006      LIKE gzpe_t.gzpe009,  #Mail address
              oojn012      LIKE gzpe_t.gzpe013,  #進階設定
              oojn007      LIKE gzpe_t.gzpe010,  #報表mail檔案格式
              ready        LIKE type_t.chr1              
                           END RECORD
   #190408-00033#2 add(e)
   
                           
   DEFINE   li_cnt         LIKE type_t.num5
   DEFINE   l_i            LIKE type_t.num5
   DEFINE   lst_maillist   base.StringTokenizer
   DEFINE   ls_mail        STRING
   DEFINE   l_ac           LIKE type_t.num5
   DEFINE   lc_gzgd000     LIKE gzgd_t.gzgd000 
   DEFINE   l_subject      VARCHAR(300)          #150709-00023#1 add
   DEFINE   l_context      STRING                #150709-00023#1 add
   DEFINE   l_source       STRING                #150709-00023#1 add    
   DEFINE   l_gzgo004      LIKE gzgo_t.gzgo004   #150803-00008#3 add
   #190408-00033#2 mark(s)
#   #150803-00008#3 add(s)
#   DEFINE l_maillist        DYNAMIC ARRAY OF RECORD   
#              oojn005      LIKE oojn_t.oojn005,  #1:收件者 2:副本 3:密件副本
#              oojn011      LIKE oojn_t.oojn011,  #聯絡對象類型                 
#              oojn003      LIKE oojn_t.oojn003,  #收件人員工編號   
#              oojn004      LIKE oojn_t.oojn004,  #姓名 
#              oojn006      LIKE oojn_t.oojn006,  #Mail address
#              oojn012      LIKE oojn_t.oojn012,  #進階設定   #190107-00075#6 add
#              oojn007      LIKE oojn_t.oojn007   #報表mail檔案格式               
#                           END RECORD   
#   #150803-00008#3 add(e)
   #190408-00033#2 mark(e)   
   #190408-00033#2 add(s)
   DEFINE l_maillist        DYNAMIC ARRAY OF RECORD
              oojn005      LIKE gzpe_t.gzpe007,  #1:收件者 2:副本 3:密件副本
              oojn011      LIKE gzpe_t.gzpe012,  #聯絡對象類型
              oojn003      LIKE gzpe_t.gzpe004,  #收件人員工編號 
              oojn004      LIKE gzpe_t.gzpe008,  #姓名 
              oojn006      LIKE gzpe_t.gzpe009,  #Mail address
              oojn012      LIKE gzpe_t.gzpe013,  #進階設定
              oojn007      LIKE gzpe_t.gzpe010   #報表mail檔案格式   
                           END RECORD   
   #190408-00033#2 add(e)
   DEFINE l_gzxa001        LIKE gzxa_t.gzxa001   #150922-00009#1 add
   DEFINE l_gzxa001_tmp    LIKE gzxa_t.gzxa001   #150922-00009#1 add
   DEFINE ls_4st_path      STRING                #170929-00002#1 add
   DEFINE l_id             LIKE oofa_t.oofa003   #聯絡對象編號   #171020-00013#1 add
   #190107-00075#6 add(s)
   DEFINE   lr_maillist    RECORD                #進階設定JSON
              recipient    STRING,
              cc           STRING,
              bcc          STRING
                           END RECORD
   #190107-00075#6 add(e)

   #170120-00031#1 add(s)
   IF g_bgjob = 'T' THEN
       CALL cl_load_4ad_interface(NULL)
   END IF
   #170120-00031#1 add(e)  
   
   
   OPEN WINDOW w_cl_gr_maillist WITH FORM cl_ap_formpath("lib","cl_gr_s02")         
#      ATTRIBUTES(STYLE="layoutwin")   #170929-00002#1 mark
   CALL cl_ui_init()

   #170929-00002#1 add(s)
   #載入畫面Style，顯示錯誤訊息
   LET ls_4st_path = os.Path.join(os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("ERP"),"cfg"),"4st"),g_lang),"style_r_tpl-H1_blue.4st")
   CALL ui.Interface.loadStyles(ls_4st_path)
   #170929-00002#1 add(e)

   CALL lr_mail.clear()
   LET li_cnt = 1

   #下拉式選單設定
   CALL cl_set_combo_scc('oojn005','157') 
#   CALL cl_set_combo_scc('oojn007','158')   #150803-00008#3 mark
   CALL cl_set_combo_scc('oojn011','1')     #150803-00008#3 add
   #150803-00008#3 add(s)
   #允許列印方式權限控管
#   CALL cl_gr_get_gzgo004("","","") RETURNING l_gzgo004   #150901-00021#1 mark
#   CALL cl_gr_format_combo(l_gzgo004,'oojn007',NULL)      #150901-00021#1 mark 
   CALL cl_rpt_get_gzgo004("","","") RETURNING l_gzgo004    #150901-00021#1 add
   CALL cl_rpt_format_combo(l_gzgo004,'oojn007',NULL)       #150901-00021#1 add 
   #150803-00008#3 add(e)
   
   #按鍵圖案
   LET gwin_curr = ui.Window.getCurrent()  #取得現行畫面
   LET gfrm_curr = gwin_curr.getForm()     #取出物件化後的畫面物件   
   CALL gfrm_curr.setElementImage("send","grsend.png")
   CALL gfrm_curr.setElementImage("cancel","grexit.png")  

   #150709-00023#1 add(s)
   #信件主旨
   IF cl_null(g_mail_subject) OR g_mail_subject = 'default' THEN
      #150709-00023#1 mark(s)
#      LET l_subject   = "\[T100",cl_getmsg("lib-00112",g_lang) CLIPPED,"\]",
#                         cl_getmsg("lib-00113",g_lang) CLIPPED," : ",cl_gr_get_progname(g_rep_code,g_lang,1) CLIPPED,"，", 
#                         cl_getmsg("lib-00114",g_lang) CLIPPED," : ", cl_get_deptname(g_site),"，",
#                         cl_getmsg("lib-00115",g_lang) CLIPPED," : ", cl_get_username(g_user) CLIPPED,"，",
#                         cl_getmsg("lib-00116",g_lang) CLIPPED," : ", g_today 
      #150709-00023#1 mark(e)
      #150709-00023#1 add(s)
      LET l_subject = cl_gr_mail_subject()
      #150709-00023#1 add(e)   
   ELSE 
      LET l_subject = g_mail_subject   
   END IF
   DISPLAY l_subject TO formonly.l_subject
   
   #信件本文
   IF cl_null(g_mail_context) OR g_mail_context = 'default' THEN
      LET l_source = cl_gr_gen_context("")      
   ELSE 
      LET l_context = g_mail_context
      LET l_source = g_mail_context      #170524-00050#1 add
   END IF
   LET l_context = l_source
   DISPLAY l_context TO formonly.l_context   
   DISPLAY l_source TO formonly.l_source
   
   IF g_schedule_d.getLength() > 0 THEN
      FOR li_cnt = 1 TO g_schedule_d.getLength()
         IF g_schedule_d[li_cnt].gzpe003 = "Z" THEN   #170618-00057#1 add
#            LET lr_mail[li_cnt].oojn011 = g_schedule_d[li_cnt].gzpe012   #150803-00008#3 add      
            LET lr_mail[li_cnt].oojn005 = g_schedule_d[li_cnt].gzpe007
            LET lr_mail[li_cnt].oojn003 = g_schedule_d[li_cnt].gzpe004
            LET lr_mail[li_cnt].oojn004 = g_schedule_d[li_cnt].gzpe008
            IF NOT cl_null(g_schedule_d[li_cnt].gzpe009) THEN
               LET lr_mail[li_cnt].oojn006 = g_schedule_d[li_cnt].gzpe009
            ELSE
#               LET lr_mail[li_cnt].oojn006 = s_aooi350_get_user_contact(g_schedule_d[li_cnt].gzpe004,'4')   #150803-00008#3 mark      
#               LET lr_mail[li_cnt].oojn006 = s_aooi350_get_cont_ppl_email(g_schedule_d[li_cnt].gzpe012,g_schedule_d[li_cnt].gzpe004)   #150803-00008#3 add   #161222-00006#1 mark
#               LET lr_mail[li_cnt].oojn006 = cl_s_aooi350_get_cont_ppl_email(g_schedule_d[li_cnt].gzpe012,g_schedule_d[li_cnt].gzpe004)                       #161222-00006#1 add   #170929-00002#1 mark
               LET lr_mail[li_cnt].oojn006 = cl_gr_get_cont_ppl_email(g_schedule_d[li_cnt].gzpe012,g_schedule_d[li_cnt].gzpe004)                               #161222-00006#1 add   #170929-00002#1 add
            END IF
            LET lr_mail[li_cnt].oojn012 = g_schedule_d[li_cnt].gzpe013   #190107-00075#6 add
            LET lr_mail[li_cnt].oojn007 = g_schedule_d[li_cnt].gzpe010
            LET lr_mail[li_cnt].ready = "Y"
         END IF   #170618-00057#1 add
      END FOR
   END IF
   #150709-00023#1 add(e)
   
   #若g_receiver_list, g_cc_list, g_bcc_list內有值, 則一起匯入列表
   IF NOT cl_null(g_receiver_list) THEN
      LET lst_maillist = base.StringTokenizer.create(g_receiver_list, ";")
      WHILE lst_maillist.hasMoreTokens()
         LET ls_mail = lst_maillist.nextToken()
         IF NOT cl_null(ls_mail) THEN
            LET lr_mail[li_cnt].oojn005 = "1"
            LET lr_mail[li_cnt].ready = "Y"
            CALL cl_gr_splite(ls_mail) RETURNING lr_mail[li_cnt].oojn006,lr_mail[li_cnt].oojn004,lr_mail[li_cnt].oojn007,l_gzxa001_tmp #排程add #150922-00009#1 add gzxa001
            IF cl_null(lr_mail[li_cnt].oojn007) THEN
               LET lr_mail[li_cnt].oojn007 = "1"
            END IF
            LET li_cnt = li_cnt + 1
         END IF
      END WHILE
   END IF
   
   IF NOT cl_null(g_cc_list) THEN
      LET lst_maillist = base.StringTokenizer.create(g_cc_list, ";")
      WHILE lst_maillist.hasMoreTokens()
         LET ls_mail = lst_maillist.nextToken()
         IF NOT cl_null(ls_mail) THEN
            LET lr_mail[li_cnt].oojn005 = "2" 
            LET lr_mail[li_cnt].ready = "Y"
            CALL cl_gr_splite(ls_mail) RETURNING lr_mail[li_cnt].oojn006,lr_mail[li_cnt].oojn004,lr_mail[li_cnt].oojn007,l_gzxa001_tmp #排程add #150922-00009#1 add gzxa001
            IF cl_null(lr_mail[li_cnt].oojn007) THEN
               LET lr_mail[li_cnt].oojn007 = "1"
            END IF
            LET li_cnt = li_cnt + 1
         END IF
      END WHILE
   END IF
  
   IF NOT cl_null(g_bcc_list) THEN
      LET lst_maillist = base.StringTokenizer.create(g_bcc_list, ";")
      WHILE lst_maillist.hasMoreTokens()
         LET ls_mail = lst_maillist.nextToken()
         IF NOT cl_null(ls_mail) THEN
            LET lr_mail[li_cnt].oojn005 = "3"
            LET lr_mail[li_cnt].ready = "Y"
            CALL cl_gr_splite(ls_mail) RETURNING lr_mail[li_cnt].oojn006,lr_mail[li_cnt].oojn004,lr_mail[li_cnt].oojn007,l_gzxa001_tmp #排程add #150922-00009#1 add gzxa001
            IF cl_null(lr_mail[li_cnt].oojn007) THEN
               LET lr_mail[li_cnt].oojn007 = "1"
            END IF
            LET li_cnt = li_cnt + 1
         END IF
      END WHILE
   END IF

   #190107-00075#6 add(s)
   IF NOT cl_null(g_adv_list) THEN
      LET lst_maillist = base.StringTokenizer.create(g_adv_list, "|")
      WHILE lst_maillist.hasMoreTokens()
         LET ls_mail = lst_maillist.nextToken()
         IF NOT cl_null(ls_mail) THEN
            LET lr_mail[li_cnt].oojn005 = "4"
            LET lr_mail[li_cnt].ready = "Y"
            LET lr_mail[li_cnt].oojn012 = ls_mail.subString(1,ls_mail.getIndexOf("&",1))
            LET lr_mail[li_cnt].oojn007 = ls_mail.subString(ls_mail.getIndexOf("&",1),ls_mail.getLength())
            IF cl_null(lr_mail[li_cnt].oojn007) THEN
               LET lr_mail[li_cnt].oojn007 = "1"
            END IF
            LET li_cnt = li_cnt + 1
         END IF
      END WHILE
   END IF
   #190107-00075#6 add(e)
   
   #匯入設定檔內的receiver, cc, bcc
   IF lr_mail.getLength() == 0 THEN 
      #150803-00008#3 add(s)
#      CALL cl_gr_get_default_maillist("","","") RETURNING l_subject,l_source,l_maillist    #150901-00021#1 mark
      CALL cl_rpt_get_default_maillist("","","") RETURNING l_subject,l_source,l_maillist    #150901-00021#1 add
      #230327-00010#1 add(s)
      #依元件自定g_mail_context優先,若空值才預設aooi998 內容
      IF NOT cl_null(g_mail_context) THEN
         LET l_source = g_mail_context      
      END IF
      #230327-00010#1 add(e)

      IF cl_null(l_subject) OR l_subject = 'default' THEN
         LET l_subject = cl_gr_mail_subject()
      END IF
      IF cl_null(l_source) OR l_source = 'default' THEN
         LET l_source = cl_gr_gen_context("")
      END IF
      LET l_context = l_source   #170928-00029#1 add

      LET li_cnt = lr_mail.getLength() + 1
      FOR l_i = 1 TO l_maillist.getLength()
         LET lr_mail[li_cnt].oojn005 = l_maillist[l_i].oojn005
         LET lr_mail[li_cnt].oojn011 = l_maillist[l_i].oojn011
         LET lr_mail[li_cnt].oojn003 = l_maillist[l_i].oojn003
         LET lr_mail[li_cnt].oojn004 = l_maillist[l_i].oojn004
         LET lr_mail[li_cnt].oojn006 = cl_gr_get_cont_ppl_email(l_maillist[l_i].oojn011,l_maillist[l_i].oojn003)         #161222-00006#1 add   #170929-00002#1 add
#         LET lr_mail[li_cnt].oojn006 = cl_s_aooi350_get_cont_ppl_email(l_maillist[l_i].oojn011,l_maillist[l_i].oojn003) #161222-00006#1 add   #170929-00002#1 mark
#         LET lr_mail[li_cnt].oojn006 = s_aooi350_get_cont_ppl_email(l_maillist[l_i].oojn011,l_maillist[l_i].oojn003)   #161222-00006#1 mark
#         LET lr_mail[li_cnt].oojn007 = cl_gr_mail_format('G',l_maillist[l_i].oojn007)   #150901-00021#1 mark
         LET lr_mail[li_cnt].oojn007 = cl_rpt_mail_format('G',l_maillist[l_i].oojn007)   #150901-00021#1 add
         LET lr_mail[li_cnt].oojn012 = l_maillist[l_i].oojn012   #190107-00075#6 add
         LET lr_mail[li_cnt].ready = 'Y'
         
         LET li_cnt = li_cnt + 1
      END FOR
      #150803-00008#3 add(e)
   END IF

   DIALOG ATTRIBUTES(UNBUFFERED)
      #150709-00023#1 add(s)
#      INPUT l_subject,l_context,l_source FROM l_subject,l_context,l_source ATTRIBUTES(WITHOUT DEFAULTS)
      INPUT l_subject,l_source FROM l_subject,l_source ATTRIBUTES(WITHOUT DEFAULTS)
#         ON CHANGE l_context
#            LET l_source = l_context
#             DISPLAY l_source TO formonly.l_source
           
         ON CHANGE l_source            #170602-00007#1 remark       
            LET l_context = l_source   #170602-00007#1 add
#            LET l_context = l_source
#            DISPLAY l_context TO formonly.l_context 
            
      END INPUT   
      #150709-00023#1 add(e)
   
      INPUT ARRAY lr_mail FROM s_detail1.*
         ATTRIBUTES(COUNT=lr_mail.getLength(), MAXCOUNT=g_max_rec,
                    INSERT ROW=FALSE, DELETE ROW=FALSE, APPEND ROW=TRUE,
                    WITHOUT DEFAULTS)

         BEFORE ROW
            LET l_ac = DIALOG.getCurrentRow("s_detail1")
            #170929-00002#1 mark(s)
#            IF cl_null(lr_mail[l_ac].oojn005) THEN LET lr_mail[l_ac].oojn005 = '1' END IF     #150803-00008#3 add
#            IF cl_null(lr_mail[l_ac].oojn011) THEN LET lr_mail[l_ac].oojn011 = '2' END IF     #150803-00008#3 add
#            IF cl_null(lr_mail[l_ac].oojn007) THEN LET lr_mail[l_ac].oojn007 = '1' END IF     #150803-00008#3 add
#            LET lr_mail[l_ac].ready = "Y"   #150803-00008#3 add
            #170929-00002#1 mark(s)
            #190107-00075#6 add(s)
            IF lr_mail[l_ac].oojn005 = "4" THEN
               CALL cl_set_comp_entry("oojn012",TRUE)
               CALL cl_set_comp_required("oojn011",FALSE)
               CALL cl_set_comp_entry("oojn003,oojn004,oojn011,oojn006",FALSE)
            ELSE
               CALL cl_set_comp_entry("oojn012",FALSE)
               CALL cl_set_comp_required("oojn011",TRUE)
               CALL cl_set_comp_entry("oojn003,oojn004,oojn011,oojn006",TRUE)
            END IF
            #190107-00075#6 add(e)

            
         BEFORE INSERT
            INITIALIZE lr_mail[l_ac].* TO NULL
            LET lr_mail[l_ac].ready = "Y"     #150803-00008#3 mark   #170929-00002#1 remark
            #170929-00002#1 add(s)
            IF cl_null(lr_mail[l_ac].oojn005) THEN LET lr_mail[l_ac].oojn005 = '1' END IF
            IF cl_null(lr_mail[l_ac].oojn011) THEN LET lr_mail[l_ac].oojn011 = '2' END IF
            IF cl_null(lr_mail[l_ac].oojn007) THEN LET lr_mail[l_ac].oojn007 = '1' END IF  
            DISPLAY lr_mail[l_ac].oojn005 TO oojn005
            DISPLAY lr_mail[l_ac].oojn011 TO oojn011
            DISPLAY lr_mail[l_ac].oojn007 TO oojn007            
            #170929-00002#1 add(e)
#            LET lr_mail[l_ac].oojn007 = "1"   #150803-00008#3 mark 
#            LET lr_mail[l_ac].oojn005 = "1"   #150803-00008#3 mark
#            LET lr_mail[l_ac].oojn011 = '2'   #150803-00008#3 mark         
            NEXT FIELD oojn005      

         #161128-00002#1 add(s)
         ON CHANGE oojn011
            LET lr_mail[l_ac].oojn003 = ''   
            LET lr_mail[l_ac].oojn004 = ''
            LET lr_mail[l_ac].oojn006 = ''   #170929-00002#1 add

            DISPLAY lr_mail[l_ac].oojn003 TO oojn003
            DISPLAY lr_mail[l_ac].oojn004 TO oojn004
            DISPLAY lr_mail[l_ac].oojn006 TO oojn006   #170929-00002#1 add
            
         BEFORE FIELD oojn003
            IF cl_null(lr_mail[l_ac].oojn011) THEN
               NEXT FIELD oojn011
            END IF
         #161128-00002#1 add(e)
            
         #170929-00002#1 add(s)
         ON CHANGE oojn003
            INITIALIZE g_qryparam.* TO NULL   #170929-00002#1 add         
            LET lr_mail[l_ac].oojn004 = ''
            LET lr_mail[l_ac].oojn006 = ''
            DISPLAY lr_mail[l_ac].oojn004 TO oojn004
            DISPLAY lr_mail[l_ac].oojn006 TO oojn006          
         #170929-00002#1 add(e)
            
         AFTER FIELD oojn003  
            #170929-00002#1 add(s)
            #檢核聯絡對象編號是否存在
            IF NOT cl_null(lr_mail[l_ac].oojn003) THEN
               #r.v參數
               INITIALIZE g_chkparam.* TO NULL
               LET g_chkparam.arg1 = lr_mail[l_ac].oojn003
               LET g_errshow = FALSE
               LET g_chkparam.err_str[1] = ""
               
               CASE lr_mail[l_ac].oojn011
                  WHEN "1"   #1.組織aooi100
                     IF NOT cl_chk_exist("v_ooef001") THEN
                        NEXT FIELD CURRENT
                     END IF
                  WHEN "2"   #2.員工aooi130
                     IF NOT cl_chk_exist("v_ooag001") THEN
                        NEXT FIELD CURRENT
                     END IF                  
                  WHEN "3"   #3.客戶/廠商apmm100/apmm800
                     IF NOT cl_chk_exist("v_pmaa001_6") THEN
                        NEXT FIELD CURRENT
                     END IF                  
                  WHEN "4"   #4.交易對象聯絡人apmm100/apmi005
                     IF NOT cl_chk_exist("v_pmaj001") THEN
                        NEXT FIELD CURRENT
                     END IF                  
                  WHEN "7"   #7.銀行anmi100
                     IF NOT cl_chk_exist("v_nmab001") THEN
                        NEXT FIELD CURRENT
                     END IF                  
                  WHEN "8"   #8.庫位arti220
                     IF NOT cl_chk_exist("v_inay001") THEN
                        NEXT FIELD CURRENT
                     END IF                  
                  WHEN "12"  #12.站點adbi253
                     IF NOT cl_chk_exist("v_dbad001") THEN
                        NEXT FIELD CURRENT
                     END IF                  
                  OTHERWISE   #aooi350
#                     LET g_chkparam.arg1 = lr_mail[l_ac].oojn011   #171020-00013#1 mark
                     LET g_chkparam.arg1 = lr_mail[l_ac].oojn003    #171020-00013#1 add
                     LET g_chkparam.arg2 = lr_mail[l_ac].oojn011    #171020-00013#1 add
                     IF NOT cl_chk_exist("v_oofa003") THEN
                        NEXT FIELD CURRENT
                     END IF                  
               END CASE
            END IF
            #170929-00002#1 add(e)
                       
#            IF NOT cl_null(lr_mail[l_ac].oojn003) THEN   #161128-00002#1 mark
               #161128-00002#1 add(s)
            IF NOT cl_null(lr_mail[l_ac].oojn003) AND NOT cl_null(lr_mail[l_ac].oojn011) THEN
               #171020-00013#1 add(s)
               IF cl_null(g_qryparam.return2) THEN
                  #交易對象聯絡人只能用聯絡人識別碼(pmaj002)當唯一key
                  IF lr_mail[l_ac].oojn011 = '4' THEN
                     SELECT pmaj002 INTO l_id FROM pmaj_t WHERE pmajent = g_enterprise AND pmajstus = 'Y' AND pmaj001 = lr_mail[l_ac].oojn003                     
                  ELSE
                     LET l_id = lr_mail[l_ac].oojn003
                  END IF
                  LET lr_mail[l_ac].oojn004 = cl_s_aooi350_get_cont_ppl_name(lr_mail[l_ac].oojn011,l_id,g_dlang)
               ELSE
                  LET lr_mail[l_ac].oojn004 = g_qryparam.return2
               END IF
               #171020-00013#1 add(s)
               #171020-00013#1 mark(s)
#               LET lr_mail[l_ac].oojn004 = cl_s_aooi350_get_cont_ppl_name(lr_mail[l_ac].oojn011,lr_mail[l_ac].oojn003,g_dlang)
#               IF cl_null(lr_mail[l_ac].oojn004) THEN
#               #161128-00002#1 add(e)
#                  INITIALIZE g_ref_fields TO NULL
#                  LET g_ref_fields[1] = lr_mail[l_ac].oojn003   
#                  CALL ap_ref_array2(g_ref_fields,"SELECT ooag011 FROM ooag_t WHERE ooagent="||g_enterprise||" AND ooag001=? ","") RETURNING g_rtn_fields
#                  LET lr_mail[l_ac].oojn004 = g_rtn_fields[1]  
#               END IF   #161128-00002#1 add
               #171020-00013#1 mark(e)
  
#               CALL s_aooi350_get_user_contact(lr_mail[l_ac].oojn003,'4') RETURNING lr_mail[l_ac].oojn006   #150803-00008#3 mark         
#               CALL s_aooi350_get_cont_ppl_email(lr_mail[l_ac].oojn011,lr_mail[l_ac].oojn003) RETURNING lr_mail[l_ac].oojn006   #150803-00008#3 add   #161222-00006#1 mark
#               CALL cl_s_aooi350_get_cont_ppl_email(lr_mail[l_ac].oojn011,lr_mail[l_ac].oojn003) RETURNING lr_mail[l_ac].oojn006                       #161222-00006#1 add   #170929-00002#1 mark
               CALL cl_gr_get_cont_ppl_email(lr_mail[l_ac].oojn011,lr_mail[l_ac].oojn003) RETURNING lr_mail[l_ac].oojn006                               #161222-00006#1 add   #170929-00002#1 add
               #170929-00002 add(s)
               
               #170929-00002 add(e)
            END IF
            
         ON ACTION controlp INFIELD oojn003
            #開窗i段
            #150803-00008#3 mark(s)
#            INITIALIZE g_qryparam.* TO NULL
#            LET g_qryparam.state = 'i'
#            LET g_qryparam.reqry = FALSE
#
#            LET g_qryparam.default1 = lr_mail[l_ac].oojn003   #給予default值
#            LET g_qryparam.default2 = "" #g_gzpb2_d[l_ac].oofa011 #全名
#
#            #給予arg
#            CALL q_ooag001_3()                                #呼叫開窗
#            LET lr_mail[l_ac].oojn003 = g_qryparam.return1    #將開窗取得的值回傳到變數
#            DISPLAY lr_mail[l_ac].oojn003 TO oojn003          #顯示到畫面上
            #150803-00008#3 mark(e)

            #150803-00008#3 add(s)
            INITIALIZE g_qryparam.* TO NULL
            LET g_qryparam.state = "i"
            LET g_qryparam.reqry = FALSE
            LET g_qryparam.default1 = lr_mail[l_ac].oojn003
            LET g_qryparam.default2 = ""
            LET g_qryparam.where = "1=1"
            LET g_qryparam.arg1 = lr_mail[l_ac].oojn011
#            CALL q_oofa003_1()                               #161128-00002#1 mark
#            LET lr_mail[l_ac].oojn003 = g_qryparam.return1   #161128-00002#1 mark
#            LET lr_mail[l_ac].oojn004 = g_qryparam.return2   #161128-00002#1 mark
            #161128-00002#1 add(s)
            CASE lr_mail[l_ac].oojn011
               WHEN "1"   #1.組織   #ooef012
                  CALL q_ooef001()
               WHEN "2"   #2.員工   #ooag002
                  CALL q_ooag001()
               WHEN "3"   #3.客戶/廠商   #pmaa027
#                  CALL q_pmaa001_5()   #170929-00002#1 mark
                  CALL q_pmaa001_32()   #170929-00002#1 add
               WHEN "4"   #4.交易對象聯絡人   #pmaj002
#                  CALL q_pmaj002_1()   #170929-00002#1 mark
                  CALL q_pmaj001()      #170929-00002#1 add
               WHEN "7"   #7.銀行
                  CALL q_nmab001()
               WHEN "8"   #8.庫位
                  CALL q_inay001()
               WHEN "12"  #12.站點   #dbad003
                  CALL q_dbad001()
               OTHERWISE
                  CALL q_oofa003_1()
            END CASE           
            
            LET lr_mail[l_ac].oojn003 = g_qryparam.return1
            #171020-00013#1 add(s)
            IF cl_null(g_qryparam.return2) THEN
               #交易對象聯絡人只能用聯絡人識別碼(pmaj002)當唯一key
               IF lr_mail[l_ac].oojn011 = '4' THEN
                  SELECT pmaj002 INTO l_id FROM pmaj_t WHERE pmajent = g_enterprise AND pmajstus = 'Y' AND pmaj001 = lr_mail[l_ac].oojn003                     
               ELSE
                  LET l_id = lr_mail[l_ac].oojn003
               END IF            
               LET lr_mail[l_ac].oojn004 = cl_s_aooi350_get_cont_ppl_name(lr_mail[l_ac].oojn011,l_id,g_dlang)
            ELSE
               LET lr_mail[l_ac].oojn004 = g_qryparam.return2
            END IF
            #171020-00013#1 add(s)
            #171020-00013#1 mark(s)
#            LET lr_mail[l_ac].oojn004 = cl_s_aooi350_get_cont_ppl_name(lr_mail[l_ac].oojn011,lr_mail[l_ac].oojn003,g_dlang)
#            IF cl_null(lr_mail[l_ac].oojn004) THEN
#               LET lr_mail[l_ac].oojn004 = g_qryparam.return2
#            END IF
            #171020-00013#1 mark(e)
            #161128-00002#1 add(e)
            
            DISPLAY lr_mail[l_ac].oojn003 TO oojn003
            DISPLAY lr_mail[l_ac].oojn004 TO oojn004
            #150803-00008#3 add(e)

            NEXT FIELD oojn003
         
         #190107-00075#6 add(s)
         ON CHANGE oojn005         
            IF lr_mail[l_ac].oojn005 = "4" THEN
               CALL cl_set_comp_entry("oojn012",TRUE)
               CALL cl_set_comp_required("oojn011",FALSE)
               CALL cl_set_comp_entry("oojn003,oojn004,oojn011,oojn006",FALSE)
               LET lr_mail[l_ac].oojn011 = ""
               LET lr_mail[l_ac].oojn003 = ""
               LET lr_mail[l_ac].oojn004 = ""          
               NEXT FIELD oojn012  
            ELSE
               LET lr_mail[l_ac].oojn012 = ""
               CALL cl_set_comp_entry("oojn012",FALSE)
               CALL cl_set_comp_required("oojn011",TRUE)
               CALL cl_set_comp_entry("oojn003,oojn004,oojn011,oojn006",TRUE)
               NEXT FIELD oojn011
            END IF

            
         BEFORE FIELD oojn012
            IF lr_mail[l_ac].oojn005 = "4" THEN
               CALL cl_set_comp_entry("oojn012",TRUE)
            ELSE
               CALL cl_set_comp_entry("oojn012",FALSE)
            END IF
            
         AFTER FIELD oojn012
      
         ON ACTION controlp INFIELD oojn012
            IF lr_mail[l_ac].oojn005 = "4" THEN
               CALL cl_schedule_set_mail(lr_mail[l_ac].oojn012) RETURNING lr_mail[l_ac].oojn012
            END IF       
         #190107-00075#6 add(e)
      END INPUT
      
      ON ACTION page_html
            LET l_context = l_source
            DISPLAY l_context TO formonly.l_context    
            
      ON IDLE g_idle_seconds
         CALL cl_on_idle()
         CONTINUE DIALOG
          
      ON ACTION send         
         ACCEPT DIALOG

      ON ACTION cancel
#         LET INT_FLAG = TRUE    #151201-00015#1 mark
         LET g_rep_flag = TRUE   #151201-00015#1 add
         EXIT DIALOG
   END DIALOG

   CLOSE WINDOW w_cl_gr_maillist

   LET g_receiver_list = NULL
   LET g_cc_list = NULL
   LET g_bcc_list = NULL
   LET g_adv_list = NULL   #190107-00075#6 add

#   IF INT_FLAG THEN           #151201-00015#1 mark
   IF g_rep_flag THEN          #151201-00015#1 add
#      LET INT_FLAG = FALSE    #151201-00015#1 mark
      LET g_rep_flag = FALSE   #151201-00015#1 add
   ELSE
      FOR li_cnt = 1 TO lr_mail.getLength()
          IF cl_null(lr_mail[li_cnt].ready) OR lr_mail[li_cnt].ready = "N" THEN
             CONTINUE FOR
          END IF
#          IF cl_null(lr_mail[li_cnt].oojn006) OR cl_null(lr_mail[li_cnt].oojn007) THEN   #190107-00075#6 mark
          IF (cl_null(lr_mail[li_cnt].oojn006) OR cl_null(lr_mail[li_cnt].oojn007))   #190107-00075#6 add
            AND (cl_null(lr_mail[li_cnt].oojn012)) THEN   #190107-00075#6 add
             CONTINUE FOR
          END IF
          
          #150922-00009#1 add(s)
          LET l_gzxa001 = ""
          SELECT gzxa001 INTO l_gzxa001
            FROM gzxa_t
            WHERE gzxaent = g_enterprise AND gzxa002 = lr_mail[li_cnt].oojn011 AND gzxa003 = lr_mail[li_cnt].oojn003
          #150922-00009#1 add(e)
          
          CASE lr_mail[li_cnt].oojn005  
             WHEN "1"
                LET g_receiver_list = g_receiver_list,
                                      lr_mail[li_cnt].oojn006 CLIPPED,":",   
                                      lr_mail[li_cnt].oojn004 CLIPPED,":",
                                      lr_mail[li_cnt].oojn007 CLIPPED,":",
                                     l_gzxa001,";"                         #150922-00009#1 add
             WHEN "2"
                LET g_cc_list = g_cc_list,
                                lr_mail[li_cnt].oojn006 CLIPPED,":",  
                                lr_mail[li_cnt].oojn004 CLIPPED,":", 
                                lr_mail[li_cnt].oojn007 CLIPPED,":",
                                l_gzxa001,";"                               #150922-00009#1 add
             WHEN "3"
                LET g_bcc_list = g_bcc_list,
                                 lr_mail[li_cnt].oojn006 CLIPPED,":",  
                                 lr_mail[li_cnt].oojn004 CLIPPED,":", 
                                 lr_mail[li_cnt].oojn007 CLIPPED,":",
                                 l_gzxa001,";"                              #150922-00009#1 add
             #190107-00075#6 add(s)
             WHEN "4"
                LET g_adv_list = g_adv_list,lr_mail[li_cnt].oojn012 CLIPPED,"&",
                                 lr_mail[li_cnt].oojn007 CLIPPED,"|"
             #190107-00075#6 add(e)
          END CASE
      END FOR
   END IF

   RETURN g_receiver_list, g_cc_list, g_bcc_list
          ,g_adv_list   #190107-00075#6 add
          ,l_subject,l_context   #150709-00023#1 add
END FUNCTION

################################################################################
# Descriptions...: GR mail的準備動作
# Memo...........:
# Usage..........: CALL cl_gr_javamail_prepare(g_mail_cnt)
#                  
# Input parameter: pi_mail_idx    郵件筆數
# Return code....: none
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_javamail_prepare(pi_mail_idx)
   DEFINE pi_mail_idx       LIKE type_t.num5
   DEFINE l_gzdel003        LIKE gzdel_t.gzdel003
   DEFINE li_i              LIKE type_t.num5
   DEFINE ls_para           STRING
   DEFINE ls_context        STRING
   DEFINE ls_pid            STRING
   DEFINE ls_context_file   STRING
   DEFINE ls_context_file1  STRING  #ls_context_file1
   DEFINE ls_cmd            STRING
   DEFINE l_xml_str         STRING  #傳給p_gr_javamail的參數
   DEFINE l_time            STRING  
   DEFINE l_chk             LIKE type_t.num5
   DEFINE l_msg             STRING
   DEFINE l_gzze003         LIKE gzze_t.gzze003
   DEFINE l_cnt             LIKE type_t.num5
   #150922-00009#1 add(s)
   DEFINE lr_maillist       DYNAMIC ARRAY OF RECORD
          address           STRING,
          username          STRING,
          account           STRING
          END RECORD
   DEFINE l_attach          STRING
   DEFINE l_gzzm011         LIKE gzzm_t.gzzm011
   #150922-00009#1 add(e)
   #170725-00086#1 add(s)
   DEFINE l_url             STRING         #附件檔案URL
   DEFINE l_tmpfile         STRING
   DEFINE l_logfile         STRING         #log檔路徑
   DEFINE lc_channel        base.Channel
   DEFINE l_log_str         STRING
   DEFINE l_file_size1      LIKE type_t.num10
   DEFINE l_file_size2      LIKE type_t.num10
   DEFINE l_time_cnt        LIKE type_t.num10
   DEFINE l_gzgd011         LIKE gzgd_t.gzgd011  #報表逾期時間
   #170725-00086#1 add(e)
   DEFINE l_context_ch      base.Channel   #190422-00009#1 add
   DEFINE l_para_data      STRING          #240130-00011#3 add
   
   IF cl_null(g_sendmail[pi_mail_idx].receiver) AND
      cl_null(g_sendmail[pi_mail_idx].cc)       AND
      cl_null(g_sendmail[pi_mail_idx].bcc)      THEN
      RETURN
   END IF

   LET l_gzdel003 = cl_gr_get_progname(g_rep_code,g_dlang,1)
   LET g_xml.recipient = g_sendmail[pi_mail_idx].receiver
   LET g_xml.cc        = g_sendmail[pi_mail_idx].cc
   LET g_xml.bcc       = g_sendmail[pi_mail_idx].bcc
   
   #判斷主旨/內文的(系統預設)
   LET l_gzze003 = g_sendmail[pi_mail_idx].subject
   SELECT COUNT(gzze003) INTO l_cnt FROM gzze_t 
    WHERE gzze001 = 'lib-00236' AND gzze003 = l_gzze003
   IF l_cnt > 0 THEN LET g_sendmail[pi_mail_idx].subject = 'default' END IF

   LET l_gzze003 = g_sendmail[pi_mail_idx].context
   SELECT COUNT(gzze003) INTO l_cnt FROM gzze_t 
    WHERE gzze001 = 'lib-00236' AND gzze003 = l_gzze003
   IF l_cnt > 0 THEN LET g_sendmail[pi_mail_idx].context = 'default' END IF   
   
   IF NOT cl_null(g_sendmail[pi_mail_idx].subject) AND g_sendmail[pi_mail_idx].subject <> "default" THEN
      LET g_xml.subject = g_sendmail[pi_mail_idx].subject
   ELSE
      #150709-00023#1 mark(s)   
#      LET g_xml.subject   = "\[T100",cl_getmsg("lib-00112",g_lang) CLIPPED,"\]",
#                            cl_getmsg("lib-00113",g_lang) CLIPPED," : ",cl_gr_get_progname(g_rep_code,g_lang,1) CLIPPED,"，", 
#                            cl_getmsg("lib-00114",g_lang) CLIPPED," : ", cl_get_deptname(g_site),"，",
#                            cl_getmsg("lib-00115",g_lang) CLIPPED," : ", cl_get_username(g_user) CLIPPED,"，",
#                            cl_getmsg("lib-00116",g_lang) CLIPPED," : ", g_today                            
      #150709-00023#1 mark(e)   
      #150709-00023#1 add(s)
      LET g_xml.subject = cl_gr_mail_subject()
      #150709-00023#1 add(e)                            
   END IF                         
   CASE g_sendmail[pi_mail_idx].format   
       WHEN "1" LET g_xml.attach = g_mail_attach,".pdf"
                LET g_gr_filename = g_gr_filename,".pdf"  
       WHEN "2" LET g_xml.attach = g_mail_attach,".xls"
                LET g_gr_filename = g_gr_filename,".xls" 
       WHEN "3" LET g_xml.attach = g_mail_attach,".xls"
                LET g_gr_filename = g_gr_filename,".xls"  
       WHEN "4" LET g_xml.attach = g_mail_attach,".xlsx"
                LET g_gr_filename = g_gr_filename,".xlsx"  
       WHEN "5" LET g_xml.attach = g_mail_attach,".xlsx"
                LET g_gr_filename = g_gr_filename,".xlsx" 
       WHEN "6" LET g_xml.attach = g_mail_attach,".rtf"
                LET g_gr_filename = g_gr_filename,".rtf"   
       WHEN "7" LET g_xml.attach = g_mail_attach,".html"
                LET g_gr_filename = g_gr_filename,".html" 
#       WHEN ""  LET g_xml.attach = ""
#                LET g_gr_filename = ""                 
      OTHERWISE #LET g_xml.attach = g_mail_attach,".pdf"   #格式預設為PDF
                #LET g_gr_filename = g_gr_filename,".pdf"  
                LET g_xml.attach = ""
                LET g_gr_filename = ""                 
   END CASE 

   #LET g_xml.attach = NULL
   
   #170725-00086#1 add(s)
   LET g_mail_resp = ""
   LET l_file_size1 = 0
   LET l_file_size2 = 0
   LET l_time_cnt = 0
   #取得報表逾期時間
   SELECT DISTINCT gzgd011 INTO l_gzgd011 FROM gzgd_t WHERE gzgd000 = g_gzgd000

   IF l_gzgd011 = 0 OR cl_null(l_gzgd011) THEN
      LET l_gzgd011 = 1200 #預設20mins
   ELSE
      LET l_gzgd011 = l_gzgd011*60 #換算為秒
   END IF
   #210329-00008#1 -S
   #LET li_i = 2   #每2秒檢查一次, 否則等待時間會過長
   LET li_i = 1    #每1秒檢查一次
   #210329-00008#1 -E
   

  #從報表主機上透過URL下載郵件附件至ERP主機$TEMP
   LET l_url = os.Path.join(os.Path.join(os.Path.join(cl_rpt_get_env_global("XTRAGRIDIP"),"Temp"),"gr"),g_gr_filename)
   LET l_tmpfile = os.Path.join(FGL_GETENV("TEMPDIR"),g_gr_filename)
   DISPLAY "l_tmpfile:",l_tmpfile
   LET l_logfile = os.Path.join(FGL_GETENV("TEMPDIR"),os.Path.basename(g_gr_filename)||".log")
   LET ls_cmd = "wget -o ",l_logfile," -P ",FGL_GETENV("TEMPDIR")," ",l_url
   LET g_attach_desc = l_url

   WHILE ( l_time_cnt < l_gzgd011)
      RUN ls_cmd
      SLEEP 1
      LET l_file_size1 = os.path.size(l_tmpfile)
      #SLEEP li_i     #210329-00008#1
      LET l_time_cnt = l_time_cnt + li_i
      LET l_file_size2 = os.path.size(l_tmpfile)
     
      IF (l_file_size1 = l_file_size2) AND ( os.Path.exists(l_tmpfile)) THEN
         EXIT WHILE
      END IF
      
      #超過20s檔案大小沒有增加時,判定下載失敗,避免卡在迴圈中太久,請檢查IIS設定
      IF l_time_cnt > 20 AND l_file_size1 <=0 THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code =  "lib-00436"   #取得報表附件失敗，請檢查報表主機IIS設定！
         LET g_errparam.extend = ""
         LET g_errparam.popup = TRUE
         CALL cl_err()      
         EXIT WHILE        
      END IF
   END WHILE
   
   #解析log內回傳值
   LET lc_channel = base.Channel.create()   #create new 物件
   CALL lc_channel.openFile(l_logfile,"r") #開啟檔案
   WHILE TRUE
      LET l_log_str = lc_channel.readLine() #整行讀入
      IF l_log_str.getIndexOf("response",1) > 0 THEN EXIT WHILE END IF
      IF lc_channel.isEof() THEN EXIT WHILE END IF     #判斷是否為最後
   END WHILE
   CALL lc_channel.close()
   LET g_mail_resp = l_log_str   #200:ok, 404:not found
   
   LET g_xml.attach = l_tmpfile
   #170725-00086#1 add(e)
   
   FOR li_i = 1 TO NUM_ARGS()
       LET ls_para = ls_para," '",ARG_VAL(li_i),"'"
   END FOR

   LET l_time = CURRENT HOUR TO FRACTION(5)   
   LET l_time = cl_replace_str(EXTEND(l_time, HOUR TO FRACTION(5)), ":", "")
   LET l_time = cl_replace_str(l_time, ".", "_")

   LET ls_pid= FGL_GETPID(),l_time 
   #信件本文
   IF NOT cl_null(g_sendmail[pi_mail_idx].context) AND g_sendmail[pi_mail_idx].context <> "default" THEN   
      LET ls_context = g_sendmail[pi_mail_idx].context    
   ELSE
      LET ls_para = cl_replace_str(ls_para,"'","\"")
      LET ls_context = cl_gr_gen_context(ls_para)
   END IF   
   #信件內文加上個資警告訊息
   #240130-00011#3 add(s)   
   #LET ls_context = ls_context,"<br>---<br><br>",cl_getmsg("lib-00219",g_lang) CLIPPED   
   #由aoos010參數決定郵件內文是否加上信件警语
   LET l_para_data = cl_get_para(g_enterprise, g_site, "E-SYS-2024")
   IF cl_null(l_para_data) THEN  #向下相容
      LET ls_context = ls_context,"<br>---<br><br>",cl_getmsg("lib-00219",g_lang) CLIPPED       
   END IF
   #240130-00011#3 add(e)
   
   IF os.Path.exists(ls_context) THEN
       LET ls_pid= FGL_GETPID(),l_time 
   END IF 

   LET ls_context_file = FGL_GETENV("TEMPDIR"),os.Path.separator(),"report_context_",ls_pid.trim(),".txt"
   #190422-00009#1 add(s)
   LET l_context_ch = base.Channel.create()
   CALL l_context_ch.openFile(ls_context_file,"a")
   CALL l_context_ch.writeLine(ls_context)
   #190422-00009#1 add(e)
   #190422-00009#1 mark(s)
#   LET ls_cmd = "echo '", ls_context,"' > ",ls_context_file
#   RUN ls_cmd
   #190422-00009#1 mark(e)
   LET g_xml.body = ls_context_file
   
   #mail資訊
   #寄信人，如果沒有指定就從azzi918取得
   IF cl_null(g_xml.sender) THEN
      CALL cl_gr_get_mailsender()
   END IF
   
   #密件副本
   #收件者
       
   #150922-00009#1 add(s)
   LET l_gzzm011 = "N"
   LET l_gzzm011 = cl_gr_get_gzzm011()
   LET g_separator_str = "{+}"   #190606-00010#1 add
   #若不加密就走原流程，若加密則收件者要拆開寄送
   IF l_gzzm011 = "N" THEN
   #150922-00009#1 add(e)
      LET l_xml_str=""
      #190606-00010#1 add(s)
      LET l_xml_str=l_xml_str,g_xml.attach,g_separator_str,g_xml.bcc ,g_separator_str,g_xml.body,g_separator_str,g_xml.cc,
          g_separator_str,g_xml.file,g_separator_str,g_xml.mailserver,g_separator_str,g_xml.passwd,g_separator_str,g_xml.recipient,
          g_separator_str,g_xml.sender,g_separator_str,g_xml.serverport,g_separator_str,g_xml.subject,g_separator_str,g_xml.user,
          g_separator_str
      LET ls_context_file1 = FGL_GETENV("TEMPDIR"),os.Path.separator(),ls_pid.trim(),".txt"
      #190606-00010#1 add(e)
   #190606-00010#1 mark(s)      
#      LET l_xml_str=l_xml_str,g_xml.attach,"|",g_xml.bcc ,"|",g_xml.body,"|",g_xml.cc,  
#          "|",g_xml.file,"|",g_xml.mailserver,"|",g_xml.passwd,"|",g_xml.recipient,     
#          "|",g_xml.sender,"|",g_xml.serverport,"|",g_xml.subject,"|",g_xml.user,       
#          "|"                                                                           
   #190606-00010#1 mark(s)
      LET ls_context_file1 = FGL_GETENV("TEMPDIR"),os.Path.separator(),ls_pid.trim(),".txt"
      #190422-00009#1 add(s)
      LET l_context_ch = base.Channel.create()  
      CALL l_context_ch.openFile(ls_context_file1,"a") 
      CALL l_context_ch.writeLine(l_xml_str)
      #190422-00009#1 add(e)
      #190422-00009#1 mark(s)
#      LET ls_cmd=""
#      LET ls_cmd = "echo '", l_xml_str,"' > ",ls_context_file1
#      RUN ls_cmd 
      #190422-00009#1 mark(e)
   
      LET g_xml.sender = ""
      LET ls_cmd=""
   
      #寄發mail
      CALL cl_gr_javamail_buffer(ls_context_file1 CLIPPED,g_rep_code)
   #150922-00009#1 add(s)
   ELSE
      LET l_attach = g_xml.attach
      
      CALL lr_maillist.clear()
      IF NOT cl_null(g_xml.recipient) THEN
         LET lr_maillist = cl_gr_get_mail_recipient_array(g_xml.recipient)
      END IF
      
      FOR li_i = 1 TO lr_maillist.getLength()
         LET g_xml.attach = cl_gr_get_enc_file(lr_maillist[li_i].address,lr_maillist[li_i].account,l_attach,g_sendmail[pi_mail_idx].format)
         #170725-00086#1 add(s)
         LET l_url = os.Path.join(os.Path.join(os.Path.join(cl_rpt_get_env_global("XTRAGRIDIP"),"Temp"),"gr"),g_gr_filename)
         LET l_tmpfile = os.Path.join(FGL_GETENV("TEMPDIR"),g_gr_filename)
         LET l_logfile = os.Path.join(FGL_GETENV("TEMPDIR"),os.Path.basename(g_gr_filename)||".log")
         LET ls_cmd = "wget -o ",l_logfile," -P ",FGL_GETENV("TEMPDIR")," ",l_url
         LET g_attach_desc = l_url
         RUN ls_cmd
         LET g_xml.attach = l_tmpfile
         #170725-00086#1 add(e)
         #190606-00010#1 mark(s)
#         LET l_xml_str=l_xml_str,g_xml.attach,"|"," ","|",g_xml.body,"|"," ",
#                       "|",g_xml.file,"|",g_xml.mailserver,"|",g_xml.passwd,"|",lr_maillist[li_i].address,":",lr_maillist[li_i].username,
#                       "|",g_xml.sender,"|",g_xml.serverport,"|",g_xml.subject,"|",g_xml.user,
#                       "|"
         #190606-00010#1 mark(e)
         #190606-00010#1 add(s)
         LET l_xml_str=l_xml_str,g_xml.attach,g_separator_str," ",g_separator_str,g_xml.body,g_separator_str," ",
                       g_separator_str,g_xml.file,g_separator_str,g_xml.mailserver,g_separator_str,g_xml.passwd,g_separator_str,lr_maillist[li_i].address,":",lr_maillist[li_i].username,
                       g_separator_str,g_xml.sender,g_separator_str,g_xml.serverport,g_separator_str,g_xml.subject,g_separator_str,g_xml.user,
                       g_separator_str
         #190606-00010#1 add(e)
         LET ls_context_file1 = FGL_GETENV("TEMPDIR"),os.Path.separator(),ls_pid.trim(),".txt"
         #190422-00009#1 add(s)
         LET l_context_ch = base.Channel.create()  
         CALL l_context_ch.openFile(ls_context_file1,"a") 
         CALL l_context_ch.writeLine(l_xml_str)
         #190422-00009#1 add(e)
         #190422-00009#1 mark(s)
#         LET ls_cmd=""
#         LET ls_cmd = "echo '", l_xml_str,"' > ",ls_context_file1
#         RUN ls_cmd
         #190422-00009#1 mark(e)
         
         LET g_xml.sender = ""
         LET ls_cmd = ""
         
         #寄發mail
         CALL cl_gr_javamail_buffer(ls_context_file1 CLIPPED,g_rep_code)
         
         IF os.Path.delete(g_xml.attach CLIPPED) THEN
         END IF
         
         SLEEP 2
      END FOR
      
      CALL lr_maillist.clear()
      IF NOT cl_null(g_xml.cc) THEN
         LET lr_maillist = cl_gr_get_mail_recipient_array(g_xml.cc)
      END IF
      
      FOR li_i = 1 TO lr_maillist.getLength()
         LET g_xml.attach = cl_gr_get_enc_file(lr_maillist[li_i].address,lr_maillist[li_i].account,l_attach,g_sendmail[pi_mail_idx].format)
         #170725-00086#1 add(s)
         LET l_url = os.Path.join(os.Path.join(os.Path.join(cl_rpt_get_env_global("XTRAGRIDIP"),"Temp"),"gr"),g_gr_filename)
         LET l_tmpfile = os.Path.join(FGL_GETENV("TEMPDIR"),g_gr_filename)
         LET l_logfile = os.Path.join(FGL_GETENV("TEMPDIR"),os.Path.basename(g_gr_filename)||".log")
         LET ls_cmd = "wget -o ",l_logfile," -P ",FGL_GETENV("TEMPDIR")," ",l_url
         LET g_attach_desc = l_url
         RUN ls_cmd
         LET g_xml.attach = l_tmpfile
         #170725-00086#1 add(e)
         #190606-00010#1 mark(s)         
#         LET l_xml_str=l_xml_str,g_xml.attach,"|"," ","|",g_xml.body,"|",lr_maillist[li_i].address,":",lr_maillist[li_i].username,
#                       "|",g_xml.file,"|",g_xml.mailserver,"|",g_xml.passwd,"|"," ",
#                       "|",g_xml.sender,"|",g_xml.serverport,"|",g_xml.subject,"|",g_xml.user,
#                       "|"
#         
#         LET ls_context_file1 = FGL_GETENV("TEMPDIR"),os.Path.separator(),ls_pid.trim(),".txt"
#         LET ls_cmd=""
#         LET ls_cmd = "echo '", l_xml_str,"' > ",ls_context_file1
#         RUN ls_cmd
         #190606-00010#1 mark(e)
         #190606-00010#1 add(s)
         LET l_xml_str=l_xml_str,g_xml.attach,g_separator_str," ",g_separator_str,g_xml.body,g_separator_str,lr_maillist[li_i].address,":",lr_maillist[li_i].username,
                       g_separator_str,g_xml.file,g_separator_str,g_xml.mailserver,g_separator_str,g_xml.passwd,g_separator_str," ",
                       g_separator_str,g_xml.sender,g_separator_str,g_xml.serverport,g_separator_str,g_xml.subject,g_separator_str,g_xml.user,
                       g_separator_str
         LET ls_context_file1 = FGL_GETENV("TEMPDIR"),os.Path.separator(),ls_pid.trim(),".txt"
         LET l_context_ch = base.Channel.create()
         CALL l_context_ch.openFile(ls_context_file1,"a")
         CALL l_context_ch.writeLine(l_xml_str) 
         #190606-00010#1 add(e)
         LET g_xml.sender = ""
         LET ls_cmd = ""
         
         #寄發mail
         CALL cl_gr_javamail_buffer(ls_context_file1 CLIPPED,g_rep_code)
         
         IF os.Path.delete(g_xml.attach CLIPPED) THEN
         END IF
         
         SLEEP 2
      END FOR
      
      CALL lr_maillist.clear()
      IF NOT cl_null(g_xml.bcc) THEN
         LET lr_maillist = cl_gr_get_mail_recipient_array(g_xml.bcc)
      END IF
      
      FOR li_i = 1 TO lr_maillist.getLength()
         LET g_xml.attach = cl_gr_get_enc_file(lr_maillist[li_i].address,lr_maillist[li_i].account,l_attach,g_sendmail[pi_mail_idx].format)
         #170725-00086#1 add(s)
         LET l_url = os.Path.join(os.Path.join(os.Path.join(cl_rpt_get_env_global("XTRAGRIDIP"),"Temp"),"gr"),g_gr_filename)
         LET l_tmpfile = os.Path.join(FGL_GETENV("TEMPDIR"),g_gr_filename)
         LET l_logfile = os.Path.join(FGL_GETENV("TEMPDIR"),os.Path.basename(g_gr_filename)||".log")
         LET ls_cmd = "wget -o ",l_logfile," -P ",FGL_GETENV("TEMPDIR")," ",l_url
         LET g_attach_desc = l_url
         RUN ls_cmd
         LET g_xml.attach = l_tmpfile
         #170725-00086#1 add(e)       
         #190606-00010#1 mark(s)         
#         LET l_xml_str=l_xml_str,g_xml.attach,"|",lr_maillist[li_i].address,":",lr_maillist[li_i].username,"|",g_xml.body,"|"," ",
#                       "|",g_xml.file,"|",g_xml.mailserver,"|",g_xml.passwd,"|"," ",
#                       "|",g_xml.sender,"|",g_xml.serverport,"|",g_xml.subject,"|",g_xml.user,
#                       "|"
#         
#         LET ls_context_file1 = FGL_GETENV("TEMPDIR"),os.Path.separator(),ls_pid.trim(),".txt"
#         LET ls_cmd=""
#         LET ls_cmd = "echo '", l_xml_str,"' > ",ls_context_file1
#         RUN ls_cmd
         #190606-00010#1 mark(e)
         #190606-00010#1 add(s)
         LET l_xml_str=l_xml_str,g_xml.attach,g_separator_str,lr_maillist[li_i].address,":",lr_maillist[li_i].username,g_separator_str,g_xml.body,g_separator_str," ",
                       g_separator_str,g_xml.file,g_separator_str,g_xml.mailserver,g_separator_str,g_xml.passwd,g_separator_str," ",
                       g_separator_str,g_xml.sender,g_separator_str,g_xml.serverport,g_separator_str,g_xml.subject,g_separator_str,g_xml.user,
                       g_separator_str
         LET ls_context_file1 = FGL_GETENV("TEMPDIR"),os.Path.separator(),ls_pid.trim(),".txt"
         LET l_context_ch = base.Channel.create()
         CALL l_context_ch.openFile(ls_context_file1,"a")
         CALL l_context_ch.writeLine(l_xml_str)
         #190606-00010#1 add(e)
         LET g_xml.sender = ""
         LET ls_cmd = ""
         
         #寄發mail
         CALL cl_gr_javamail_buffer(ls_context_file1 CLIPPED,g_rep_code)
         
         IF os.Path.delete(g_xml.attach CLIPPED) THEN
         END IF
         
         SLEEP 2
      END FOR
   END IF
   #150922-00009#1 add(e)
END FUNCTION

################################################################################
# Descriptions...: 設定GR mail的郵件本文
# Memo...........:
# Usage..........: LET ls_context = cl_gr_gen_context(ls_para)
#                  
# Input parameter: ls_para        
# Return code....: ls_context     郵件本文
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_gen_context(ps_para)
   DEFINE ps_para     STRING
   DEFINE ls_context  STRING
   DEFINE l_maillogo  STRING
   #170602-00007#1 add(s)
   DEFINE ls_date     STRING   
   DEFINE ls_time     DATETIME HOUR TO SECOND
   DEFINE l_time      DATETIME YEAR TO SECOND
   #170602-00007#1 add(e)

   LET l_maillogo = os.Path.join(os.Path.join(FGL_GETENV("FGLASIP"),"pic"),"T100_TW.png")  
   
   #170426-00038#1 add(s)
   IF NOT cl_null(g_xml.body) OR NOT cl_null(g_schedule_d[1].gzpe006) THEN
      IF cl_null(g_xml.body) THEN
         LET ls_context = g_schedule_d[1].gzpe006
      ELSE
         LET ls_context = g_xml.body      
      END IF      
   ELSE
   #170426-00038#1 add(e)
   #170602-00007#1 add(s)
   LET l_time = cl_get_current()
   LET ls_date = DATE(l_time) USING g_date_fmt
   LET ls_time = l_time
   #170602-00007#1 add(e)
      LET ls_context =  "<html><head><title>jmail</title><style type=\"text/css\">.table {font-family: \"微軟正黑體\";background-color:#f7f7f7;}",
                        ".table_title {font-family: \"微軟正黑體\";color: #940f14;text-align: right;background-color:#f7f7f7;}",
                        ".table_bottom {font-family: \"微軟正黑體\";color: #0a0d5a;text-align: right;background-color:#f7f7f7;",
                        "-webkit-border-bottom-right-radius: 5px;-webkit-border-bottom-left-radius: 5px;-moz-border-radius-bottomright: 5px;-moz-border-radius-bottomleft: 5px;",
                        "border-bottom-right-radius: 5px;border-bottom-left-radius: 5px;}",
                        ".massage_title{padding-left:20px;padding-top:10px;background-color:#f7f7f7;font-family: \"微軟正黑體\";",
	                    "-webkit-border-top-left-radius: 5px;-webkit-border-top-right-radius: 5px;-moz-border-radius-topleft: 5px;-moz-border-radius-topright: 5px;",
                        "border-top-left-radius: 5px;border-top-right-radius: 5px;border-bottom:double;border-bottom-color:#b31a20;}</style></head>",                        
                        "<body><table class=\"tables\" cellspacing=\"0\" cellpadding=\"0\"><tr><td><table width=\"800\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\">",    #170725-00086#1 add
#                        "<body><table class=\"tables\" cellspacing=\"0\" cellpadding=\"0\"><tr><td><table width=\"543\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\">",   #170725-00086#1 mark
                        #系統通知
                        "<tr><td height=\"60\" colspan=\"2\" class=\"massage_title\" ><p class=\"table\"><strong> <img src=",l_maillogo," width=\"100\" height=\"21\" alt=\"T100 logo\"> ",cl_getmsg("lib-00112",g_lang) CLIPPED," </strong></p></td></tr>",
                        "<tr><td height=\"10\" colspan=\"2\" valign=\"center\" class=\"table_title\"></td></tr>",
                        #報表名稱
                        "<tr><td width=\"141\" height=\"40\" valign=\"center\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00113",g_lang) CLIPPED,"</strong>：</p></td>"   #161227-00007#1 add
                        #"<tr><td width=\"141\" height=\"40\" valign=\"center\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00113",g_lang) CLIPPED,"</strong>：</p></td>",  #161227-00007#1 mark
                        #"<td width=\"394\" valign=\"center\" class=\"table\"><p><span >",cl_gr_get_progname(g_rep_code,g_lang,2) CLIPPED,"</span></p></td></tr>",   #160921-00007#3 mark
                        #"<td width=\"394\" valign=\"center\" class=\"table\"><p><span >",cl_get_progname(g_prog,g_lang,2) CLIPPED,"</span></p></td></tr>",           #160921-00007#3 add   #161227-00007#1 mark
                        #161227-00007#1 add -(s)
                        IF NOT cl_null(g_rep_code_desc) THEN
#                           LET ls_context = ls_context,"<td width=\"394\" valign=\"center\" class=\"table\"><p><span >",g_rep_code_desc CLIPPED,"</span></p></td></tr>"   #170725-00086#1 mark
                           LET ls_context = ls_context,"<td width=\"600\" valign=\"center\" class=\"table\"><p><span >",g_rep_code_desc CLIPPED,"</span></p></td></tr>"    #170725-00086#1 add
                        ELSE
#                           LET ls_context = ls_context,"<td width=\"394\" valign=\"center\" class=\"table\"><p><span >",cl_gr_get_progname(g_rep_code,g_lang,2) CLIPPED,"</span></p></td></tr>"   #170725-00086#1 mark
                           LET ls_context = ls_context,"<td width=\"600\" valign=\"center\" class=\"table\"><p><span >",cl_gr_get_progname(g_rep_code,g_lang,2) CLIPPED,"</span></p></td></tr>"    #170725-00086#1 add
                        END IF
                        #製表人員
                        LET ls_context = ls_context, "<tr><td width=\"141\" valign=\"center\" height=\"40\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00115",g_lang) CLIPPED,"</strong>：</p></td>",
#                                         "<td width=\"394\" valign=\"center\" class=\"table\"><p><span >",g_user CLIPPED,".",cl_get_username(g_user) CLIPPED,"</span></p></td></tr>",   #170725-00086#1 mark
                                         "<td width=\"600\" valign=\"center\" class=\"table\"><p><span >",g_user CLIPPED,".",cl_get_username(g_user) CLIPPED,"</span></p></td></tr>",    #170725-00086#1 add
                        #執行時間
                                          "<tr><td width=\"141\" height=\"40\" valign=\"center\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00117",g_lang) CLIPPED,"：</strong></p></td>",
                                          "<td width=\"600\" valign=\"center\" class=\"table\"><p><span>",ls_date," ",ls_time,"</span></p></td></tr>"                            #170602-00007#1 add    #170725-00086#1 add
#                                          "<td width=\"394\" valign=\"center\" class=\"table\"><p><span>",ls_date," ",ls_time,"</span></p></td></tr>"                            #170602-00007#1 add   #170725-00086#1 mark
#                                          "<td width=\"394\" valign=\"center\" class=\"table\"><p><span>",TODAY," ",TIME,"</span></p></td></tr>"                                #170602-00007#1 mark
                        #161227-00007#1 add -(e)  
                        #161227-00007#1 mark -(s)
                        #製表(發送)人員
                        #"<tr><td width=\"141\" valign=\"center\" height=\"40\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00115",g_lang) CLIPPED,"</strong>：</p></td>", #161227-00007#1 mark
                        #"<td width=\"394\" valign=\"center\" class=\"table\"><p><span >",g_user CLIPPED,".",cl_get_username(g_user) CLIPPED,"</span></p></td></tr>",
                        ##執行時間
                        #"<tr><td width=\"141\" height=\"40\" valign=\"center\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00117",g_lang) CLIPPED,"：</strong></p></td>",
                        #"<td width=\"394\" valign=\"center\" class=\"table\"><p><span>",TODAY," ",TIME,"</span></p></td></tr>"
                        #161227-00007#1 mark -(e)
                        #列印條件 
                        IF g_bgjob!=1 AND NOT cl_null(ps_para) THEN
                           LET ls_context=ls_context,"<tr><td width=\"141\" height=\"40\" valign=\"center\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00118",g_lang) CLIPPED,"：</strong></p></td>",
                                                     "<td width=\"600\" valign=\"center\" class=\"table\"><p><span>",ps_para,"</span></p></td></tr>"    #170725-00086#1 add
#                                                     "<td width=\"394\" valign=\"center\" class=\"table\"><p><span>",ps_para,"</span></p></td></tr>"   #170725-00086#1 mark
                        END IF
                     #執行結果
#                     LET ls_context=ls_context,"<tr><td height=\"40\" valign=\"center\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00119",g_lang) CLIPPED,"：</strong></p></td>",   #170725-00086#1 mark
#                                               "<td valign=\"center\" class=\"table\" ><p><span >",cl_getmsg("lib-00197",g_lang) CLIPPED,":",g_gr_filename,"</span></p></td></tr>",        #170725-00086#1 mark
                     #170725-00086#1 add(s)
                     LET ls_context=ls_context,"<tr><td height=\"40\" valign=\"center\" class=\"table_title\"><p><strong>",cl_getmsg("lib-00119",g_lang) CLIPPED,"：</strong></p></td>"
                     #附件說明若是空的，先寫入g_attach_desc，方便開窗寄送郵件時取替代
                     LET ls_context=ls_context,"<td valign=\"center\" class=\"table\" ><p><span >",cl_getmsg("lib-00197",g_lang) CLIPPED,":","g_attach_desc","</span></p></td></tr>"                    
                     LET ls_context=ls_context,              
                     #170725-00086#1 add(e)
                                               "<tr><td height=\"20\" colspan=\"2\" valign=\"center\" class=\"table_bottom\"></td></tr>",
                                               "</table></td></tr></table></body></html>"
   END IF   #170426-00038#1 add

   RETURN ls_context
END FUNCTION

################################################################################
# Descriptions...: 取得GR歷史報表的檔案路徑,且根據檔案存在與否以及產生時間，選擇欲覆蓋檔案
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_code         傳入程式代號
# Return code....: l_name         返回檔案完整名稱
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_savename(p_code,p_gzgd000)
   DEFINE p_gzgd000           LIKE gzgd_t.gzgd000
   DEFINE p_code              LIKE gzde_t.gzde001,
          l_name              LIKE type_t.chr100,    #檔案名稱長度
          l_temp_path         STRING,    #用來儲存路徑及檔案名稱
          l_time_ori          STRING,    #儲存檔案最後修改時間
          l_time_yymmdd       STRING,    #儲存修改時間之年月日
          l_time_hhmmss       STRING,    #儲存修改時間之時分秒
          l_oldest            SMALLINT   #指出最早產生的檔案是第幾份 
   DEFINE l_yymmdd_comp       DYNAMIC ARRAY OF INTEGER   #儲存年月日
   DEFINE l_oojo004           LIKE oojo_t.oojo004        #GR歷史報表留存設定檔-留存份數
   DEFINE l_hhmmss_comp       DYNAMIC ARRAY OF INTEGER   #儲存時分秒
   DEFINE i                   INTEGER    #FOR迴圈計數用 
   DEFINE i_min               INTEGER    #FOR迴圈計數用 
   DEFINE i_max               INTEGER    #FOR迴圈計數用 
   DEFINE i_str               STRING     #儲存檔案目前份數,將整數轉為字串用
   DEFINE l_filename          STRING
   #DEFINE l_gcw               RECORD LIKE gcw_file.*  #CR報表檔名設定檔 #待規劃
   DEFINE l_min,l_max         LIKE type_t.num5
   DEFINE l_str               STRING
   DEFINE l_item_name         STRING
   DEFINE l_gzgd              RECORD LIKE gzgd_t.*
   DEFINE l_gzxb001           LIKE gzxb_t.gzxb001  #使用者名稱
   DEFINE l_gzxb003           LIKE gzxb_t.gzxb003  #權限類別    
   DEFINE l_n,l_diff          LIKE type_t.num5
   DEFINE l_time              LIKE type_t.chr8
   DEFINE l_oojp              RECORD LIKE oojp_t.* #歷史報表記錄檔
   DEFINE l_sql               STRING
   DEFINE l_i                 LIKE type_t.num5
   DEFINE l_timestamp         LIKE type_t.chr50

  
   #先判斷是否維護 p_gr_history，取得留存份數  
   CALL cl_gr_chk_oojo004() RETURNING l_oojo004
   IF NOT cl_null(l_oojo004) THEN
      LET i_min = 0
      CASE l_oojo004
#         WHEN "0"  #依系統設定
#            #CASE g_aza.aza49  #140609 mark
#            #  WHEN "1" LET l_max = 10
#            #  WHEN "2" LET l_max = 20
#            #  WHEN "3" LET l_max = 30
#            #  WHEN "4" LET l_max = 40
#            # WHEN "5" LET l_max = 50
#            #END CASE
#            LET l_max = 50
         WHEN "1"  #00~09 
            LET l_max = 10
         WHEN "2"  #00~19 
            LET l_max = 20
         WHEN "3"  #00~29 
            LET l_max = 30
         WHEN "4"  #00~39 
            LET l_max = 40
         WHEN "5"  #00~49 
            LET l_max = 50
      END CASE
      #檢查是否有自訂命名，若無則依標準命名
      SELECT * INTO l_gzgd.* FROM gzgd_t 
       WHERE gzgd000 = p_gzgd000 
      #140609 mark 
#      INITIALIZE l_gcw.* TO NULL 
      LET l_sql = "SELECT * FROM oojp_t ",
                  " WHERE oojpent = '",g_enterprise,"'",
                  "   AND oojpsite = '",g_site,"'",
                  "   AND oojp001 = '",g_prog,"'",
                  "   AND oojp002 = '",g_user,"'",
                  "   AND oojp006 = 'Y'",
                  " ORDER BY oojp004"
      PREPARE oojp_pr1 FROM l_sql
      DECLARE oojp_cur1 CURSOR FOR oojp_pr1
      LET l_sql = "SELECT * FROM oojp_t ",
                  " WHERE oojpent = '",g_enterprise,"'",
                  "   AND oojpsite = '",g_site,"'",
                  "   AND oojp001 = '",g_prog,"'",
                  "   AND oojp002 = '",g_user,"'",
                  "   AND oojp006 = 'Y'",
                  " ORDER BY oojp004 DESC"
      PREPARE oojp_pr2 FROM l_sql
      DECLARE oojp_cur2 CURSOR FOR oojp_pr2
      #找目前有效的歷史報表有幾筆
#      SELECT COUNT(*) INTO l_n FROM oojp_t   #170525-00105#1 mark
      SELECT COUNT(1) INTO l_n FROM oojp_t    #170525-00105#1 add
         WHERE oojpent = g_enterprise
           AND oojpsite = g_site
           AND oojp001 = g_prog
           AND oojp002 = g_user
           AND oojp006 = 'Y'

#      #有設定自訂檔名(gcw) 報表檔名第一段~第六段
#      IF (NOT cl_null(l_gcw.gcw05)) OR (NOT cl_null(l_gcw.gcw06)) OR (NOT cl_null(l_gcw.gcw07)) OR (NOT cl_null(l_gcw.gcw08)) OR
#         (NOT cl_null(l_gcw.gcw09)) OR (NOT cl_null(l_gcw.gcw10)) THEN
#         SELECT zx02,zx04 INTO l_zx02,l_zx04 FROM zx_file WHERE zx01 = g_user   #使用者名稱/權限類別
#         SELECT gaz03,gaz06 INTO l_gaz03,l_gaz06 FROM gaz_file   #多語系程式名稱/報表抬頭
#          WHERE gaz01=g_prog AND gaz02=g_rlang
#         IF cl_null(l_gaz06) THEN
#            LET l_gaz06 = l_gaz03 CLIPPED
#         END IF 
#         LET l_time=TIME
#         LET l_item_name = p_code CLIPPED,"|",l_gaz06 CLIPPED,"|",g_plant CLIPPED,"|",g_dbs CLIPPED,"|",g_user CLIPPED,"|",l_zx02 CLIPPED,"|",g_pdate CLIPPED,"|",l_time CLIPPED       
#              
#         CALL cl_xg_filename(l_item_name,l_gcw.gcw05) RETURNING l_str  #CR報表檔名設定
#         LET l_filename = l_filename CLIPPED,l_str CLIPPED
#        CALL cl_xg_filename(l_item_name,l_gcw.gcw06) RETURNING l_str  #CR報表檔名設定
#         LET l_filename = l_filename CLIPPED,l_str CLIPPED
#         CALL cl_xg_filename(l_item_name,l_gcw.gcw07) RETURNING l_str  #CR報表檔名設定
#         LET l_filename = l_filename CLIPPED,l_str CLIPPED
#         CALL cl_xg_filename(l_item_name,l_gcw.gcw08) RETURNING l_str  #CR報表檔名設定
#         LET l_filename = l_filename CLIPPED,l_str CLIPPED
#         CALL cl_xg_filename(l_item_name,l_gcw.gcw09) RETURNING l_str  #CR報表檔名設定
#         LET l_filename = l_filename CLIPPED,l_str CLIPPED
#         CALL cl_xg_filename(l_item_name,l_gcw.gcw10) RETURNING l_str  #CR報表檔名設定
#         LET l_filename = l_filename CLIPPED,l_str CLIPPED
#          #刪除最後一個分隔符號
#         LET l_str = l_filename CLIPPED
#         LET l_str = l_str.substring(1,l_str.getLength()-1) CLIPPED
#         LET l_filename = l_str CLIPPED
#      ELSE  #不自訂命名，依標準命名
         LET l_filename = g_enterprise
         LET l_filename = l_filename CLIPPED, "_", g_site CLIPPED, "_", g_prog CLIPPED, "_", g_user CLIPPED   #組合檔名
#      END IF

      LET l_timestamp = cl_get_timestamp()
      LET l_timestamp = cl_replace_str(l_timestamp,"-","")
      LET l_timestamp = cl_replace_str(l_timestamp," ","")
      LET l_timestamp = cl_replace_str(l_timestamp,":","")
      LET l_timestamp = cl_replace_str(l_timestamp,".","")

      IF l_n >= l_max THEN #數量超出最大留存量
         LET l_diff = l_n - l_max + 1
         LET l_i = 0
         #把超出數量的歷史報表檔案變更為無效
          FOREACH oojp_cur1 INTO l_oojp.*
              LET l_i = l_i + 1
              IF l_i <= l_diff THEN
                 UPDATE oojp_t SET oojp006 = 'N'
                  WHERE oojpent = l_oojp.oojpent
                    AND oojpsite = l_oojp.oojpsite
                    AND oojp001 = l_oojp.oojp001
                    AND oojp002 = l_oojp.oojp002
                    AND oojp004 = l_oojp.oojp004
              END IF
          END FOREACH
         #取最後一筆的流水號
          IF LENGTH(l_oojp.oojp005) > 6 THEN
             LET i_str = l_oojp.oojp005[LENGTH(l_oojp.oojp005)-6,LENGTH(l_oojp.oojp005)-5] #取xxr.pdf的xx
             LET i_str = (i_str + 1 ) USING '&&'
             IF i_str >= l_max THEN LET i_str = '00' END IF
          ELSE
             LET i_str = '00'
          END IF
          #LET l_filename = l_filename CLIPPED,"_",i_str, "r.pdf"
      ELSE
         IF l_n = 0 THEN
            #LET l_filename = l_filename CLIPPED,"_00r.pdf"
         ELSE
            FOREACH oojp_cur2 INTO l_oojp.*
                EXIT FOREACH
            END FOREACH
            LET i_str = l_oojp.oojp005[LENGTH(l_oojp.oojp005)-6,LENGTH(l_oojp.oojp005)-5] #取xxr.pdf的xx
            LET i_str = (i_str + 1 ) USING '&&'
            #LET l_filename = l_filename CLIPPED,"_",i_str, "r.pdf"
         END IF
      END IF
   ELSE
      RETURN l_filename
   END IF

   LET l_filename = l_filename CLIPPED,"_",l_timestamp, ".pdf"

   RETURN l_filename

END FUNCTION

################################################################################
# Descriptions...: 修改控制GR歷史報表產生的狀態碼
# Memo...........:
# Usage..........: CALL cl_gr_history()
#                  
# Input parameter: void
# Return code....: none
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_history()

   IF g_gr_times = 2 THEN
      LET g_save_flag = "S"
      LET g_gr_times = 0
      #190321-00031#1 mark(s)
#      #190117-00039#1 add -(s)
#      IF cl_gr_get_user_gzxa020() ="N" THEN  #當沒有設定個人的列印視窗是保留時，預設是使用完關窗
#         LET g_rep_flag = TRUE    #當已做完歷史報表留存時要跳離WHILE
#      END IF
#      #190117-00039#1 add -(e)
      #190321-00031#1 mark(e)
      #當不保留視窗，又要儲存歷史報表，需在進入handler判斷歷史檔案檔名
      LET g_rep_flag = NULL   #190321-00031#1 add
   ELSE
      IF cl_null(g_savefilename) THEN
         LET g_rep_flag = NULL   #190321-00031#1 add
         LET g_save_flag = "Y"         #歷史報表每次查詢只產生一次
         LET g_gr_times = 2
      END IF
   END IF
END FUNCTION

################################################################################
# Descriptions...: 設定GR分散式模式
# Memo...........:
# Usage..........: CALL cl_gr_setDistributed(l_gzgd001,l_gzgd003,l_gzgd007)
#                  
# Input parameter: l_gzgd001      報表元件代號
#                : l_gzgd003      客製否
#                : l_gzgd007      樣板名稱(4rp)
# Return code....: none
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_setDistributed(l_gzgd001,l_gzgd003,l_gzgd007)
   DEFINE l_gzgd001           LIKE gzgd_t.gzgd001  #程式代號
   DEFINE l_gzgd003           LIKE gzgd_t.gzgd003  #客製否
   DEFINE l_gzgd007           LIKE gzgd_t.gzgd007  #樣板名稱
   DEFINE l_prefix            STRING   #190408-00033#2 add
         
   #分散模式處理
      #LET l_greport = fgl_GETENV("WINGREPORT")
#      CALL fgl_report_configureDistributedProcessing(FGL_GETENV("GRIP"),FGL_GETENV("WINGREPORT"))                                              #150916-00002#1 mark    
#      CALL fgl_report_configureDistributedEnvironment(FGL_GETENV("WINFGLDIR"),FGL_GETENV("WINFGLPROFILE"),NULL,NULL)                           #150916-00002#1 mark
      CALL fgl_report_configureDistributedProcessing(cl_rpt_get_env_global("GRIP"),cl_rpt_get_env_global("WINGREPORT"))                         #150916-00002#1 add              
      IF g_erpid = "T100ERP" THEN                                                                                                               #160608-00012#1 add              
         CALL fgl_report_configureDistributedEnvironment(cl_rpt_get_env_global("WINFGLDIR"),cl_rpt_get_env_global("WINFGLPROFILE"),NULL,NULL)  #150916-00002#1 add   #170406-00030#1 mark   #170912-00018#1 remark
#         CALL fgl_report_configureDistributedEnvironment(cl_rpt_get_env_global("WINFGLDIR"),cl_rpt_get_env_global("WINFGLPROFILE"),g_4rp_dir,NULL)                    #170406-00030#1 add  #170912-00018#1 mark
      ELSE                                                                                                                                      #160608-00012#1 add  
         CALL fgl_report_configureDistributedEnvironment(NULL,NULL,NULL,NULL)                                                                   #160608-00012#1 add
      END IF                                                                                                                                    #160608-00012#1 add
     
      #190408-00033#2 add(s)
      #T100V3.0 GREHTTPD 匯出檔案
      
      ##250224-00011#1(s)
      &ifdef g3
         LET l_prefix = "http://",cl_rpt_get_env_global("GRIP"),":",cl_rpt_get_env_global("WINGREHTTPDPORT")
         CALL fgl_report_configureDistributedURLPrefix(l_prefix)
      &endif
      ##250224-00011#1(e)

      
      &ifndef v1
         LET l_prefix = "http://",cl_rpt_get_env_global("GRIP"),":",cl_rpt_get_env_global("WINGREHTTPDPORT") 
         CALL fgl_report_configureDistributedURLPrefix(l_prefix)
      &endif
      #190408-00033#2 add(e)      
END FUNCTION

################################################################################
# Descriptions...: 判斷是否為Superuser
# Memo...........:
# Usage..........: CALL cl_gr_chk_tiptopgroup() 
#                  RETURNING l_chk
# Input parameter: void
# Return code....: l_chk      
# Date & Author..: 
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_chk_tiptopgroup()
   DEFINE l_cmd        STRING
   DEFINE l_channel    base.Channel
   DEFINE l_cnt        LIKE type_t.num5
   DEFINE l_result     STRING
   DEFINE l_chk        LIKE type_t.chr1
   DEFINE l_roles      STRING                #160608-00012#1 add   

   #160608-00012#1 add(s)
   #判斷是否為superuser
   LET l_roles = cl_replace_str(g_roles,",","','")
   LET l_roles = "('",l_roles,"')"   
   LET l_cmd = " SELECT COUNT(1) FROM gzya_t WHERE gzya001 IN ",l_roles,
               "    AND gzyastus = 'Y' AND gzyaent = '",g_enterprise,"' AND gzya003 = 'Y' "
   PREPARE cl_gr_superuser_pre FROM l_cmd
   EXECUTE cl_gr_superuser_pre INTO l_cnt
   
#   IF l_cnt > 0 THEN LET l_chk = "Y" END IF   #170301-00053#2 mark
   #160608-00012#1 add(e)
  
   #170301-00053#2 add(s)   
   IF l_cnt > 0 THEN 
      LET l_chk = "Y" 
   ELSE
      LET l_chk = "N" 
   END IF
   #170301-00053#2 add(e)   

   #160608-00012#1 mark(s)
#   LET l_cmd = "groups"
#   LET l_channel = base.Channel.create()
#   CALL l_channel.openPipe(l_cmd,"r")
#   CALL l_channel.setDelimiter("")
#   WHILE l_channel.read(l_result)
#      LET l_cnt= l_result.getIndexOf("tiptop",1)
#      IF l_cnt > 0 THEN
#         LET l_chk = 'Y'
#      ELSE
#         LET l_chk = 'N'
#      END IF
#   END WHILE
#   CALL l_channel.close()
   #160608-00012#1 mark(e)

   RETURN l_chk
END FUNCTION

################################################################################
# Descriptions...: 取得簽核圖檔位置
# Memo...........:
# Usage..........: CALL cl_gr_get_apr_loc(p_gzgd009,p_gzgd010)               
#                  RETURNING g_grPageHeader.apr_loc
# Input parameter: p_gzgd009      是否列印簽核
#                : p_gzgd010      列印簽核位置
# Return code....: STRING         列印簽核位置(0:不列印簽核/ 1:頁尾/ 2:表尾)
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_apr_loc(p_gzgd009,p_gzgd010)
   DEFINE p_gzgd009  LIKE gzgd_t.gzgd009
   DEFINE p_gzgd010  LIKE gzgd_t.gzgd010
   DEFINE l_loc    STRING

    #列印簽核位置(0:不列印簽核/ 1:頁尾/ 2:表尾)
   IF cl_null(p_gzgd009) THEN
      LET p_gzgd009 ="N"   #預設不列印簽核
   END IF
   IF cl_null(p_gzgd010) THEN
      LET p_gzgd010 ="1"   #預設列印簽核未置於頁尾
   END IF
   IF p_gzgd009 = "Y" THEN
      LET l_loc = p_gzgd010 CLIPPED
      LET g_grPageHeader.sign_show = "Y"
   ELSE
      LET l_loc = "0"
      LET g_grPageHeader.sign_show = "N"
   END IF

   RETURN l_loc
END FUNCTION

################################################################################
# Descriptions...: 讀取4rp紙張尺寸，判斷是否為寬度大於高度的紙張格式
# Memo...........:
# Usage..........: 
#                  
# Input parameter: p_path         樣板檔路徑
# Return code....: TRUE/FALSE
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_chk_landscape_paper(p_path)
   DEFINE p_path         STRING
   DEFINE l_rootnode     om.DomNode
   DEFINE l_curnode      om.DomNode
   DEFINE l_parnode      om.DomNode
   DEFINE l_nodes        om.NodeList 
   DEFINE l_doc          om.DomDocument
   DEFINE l_paperwidth   STRING
   DEFINE l_paperlength  STRING
   DEFINE l_tagname      STRING
   DEFINE l_i            LIKE type_t.num5
   DEFINE l_width        LIKE type_t.num15_3
   DEFINE l_length       LIKE type_t.num15_3
   DEFINE l_unit         LIKE type_t.chr1
   DEFINE l_gzgi007      LIKE gzgi_t.gzgi007  #指定列印方向

   LET l_doc = om.DomDocument.createFromXmlFile(p_path)
   IF l_doc IS NOT NULL THEN
      LET l_rootnode = l_doc.getDocumentElement()
      IF l_rootnode IS NOT NULL THEN     
         #抓取紙張大小   
         LET l_tagname = "report:Settings"
         LET l_nodes = l_rootnode.selectByTagName(l_tagname)
         FOR l_i = 1 TO l_nodes.getLength()
            LET l_curnode=l_nodes.item(l_i)
            EXIT FOR
         END FOR
         LET l_paperwidth =l_curnode.getAttribute("RWPageWidth")
         LET l_paperlength=l_curnode.getAttribute("RWPageLength")

         CALL cl_gr_rm_unit(l_paperwidth) RETURNING l_width, l_unit
         CALL cl_gr_rm_unit(l_paperlength) RETURNING l_length, l_unit           
           
         SELECT gzgi007 INTO l_gzgi007 FROM gzgi_t
          WHERE gzgi004 = l_width AND gzgi005 = l_length AND gzgi006 = l_unit
         IF l_gzgi007 = "1" THEN 
            RETURN TRUE
         ELSE
            RETURN FALSE
         END IF         
      END IF 
   END IF
   RETURN FALSE
END FUNCTION

################################################################################
# Descriptions...: 移除長度單位字串
# Memo...........:
# Usage..........: CALL cl_gr_rm_unit(l_paperwidth) 
#                  RETURNING l_width, l_unit
# Input parameter: p_position     長度
# Return code....: l_val          轉換後的長度
#                  l_unit         單位 
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_rm_unit(p_position)
   DEFINE p_position   STRING
   DEFINE l_strbuf     base.StringBuffer
   DEFINE l_str        STRING
   DEFINE l_val        LIKE type_t.num15_3
   DEFINE l_unit       LIKE type_t.chr1

   LET l_val = -1
#   IF p_position IS NOT NULL THEN    #160108-00011#1 mark
   IF NOT cl_null(p_position) THEN    #160108-00011#1 add
      LET l_strbuf = base.StringBuffer.create()
        
      IF p_position.getIndexOf("cm",1) > 0 THEN
         CALL l_strbuf.append(p_position)
         CALL l_strbuf.replace("cm","",0)
         LET l_str = l_strbuf.toString()
         LET l_val = l_str
         LET l_unit = "C"
      ELSE
         IF p_position.getIndexOf("in",1) > 0
            OR p_position.getIndexOf("inch",1) > 0
            THEN
            CALL l_strbuf.append(p_position)
            CALL l_strbuf.replace("inch","",0)
            CALL l_strbuf.replace("in","",0)
            LET l_str = l_strbuf.toString()
            LET l_val = l_str
            #LET l_val = l_val * 2.54 #140619 不轉換
            LET l_unit = "I"
         ELSE
            #紙張單位是pixel
            TRY
               LET l_val = p_position
               LET l_val = l_val * 2.54 / 72
            CATCH
               LET l_val = -1
            END TRY
         END IF 
      END IF
   END IF
   RETURN l_val, l_unit
END FUNCTION

################################################################################
# Descriptions...: 取得javamail附件檔名
# Memo...........:
# Usage..........: LET g_gr_filename = cl_gr_jmail_filename()
#                  
# Input parameter: void
# Return code....: STRING         附件檔名
# Date & Author..: 
# Modify.........:
################################################################################
PRIVATE FUNCTION cl_gr_jmail_filename()
   DEFINE ls_filename   STRING
   DEFINE ls_date       STRING
   DEFINE ls_time       STRING
   DEFINE l_time        DATETIME YEAR TO FRACTION(5)

   #171206-00020#1 add(s)
   #優先使用報表程式中指定的檔名
   IF NOT cl_null(g_rep_filename) THEN
      LET ls_filename = g_rep_filename
   ELSE
   #171206-00020#1 add(e)
#      LET l_time = cl_get_current()   #180920-00021#1 mark
      LET l_time = cl_get_timestamp()   #180920-00021#1 add
      LET ls_date = DATE(l_time) USING "yyyymmdd"
      LET ls_time = cl_replace_str(EXTEND(l_time, HOUR TO FRACTION(5)), ":", "")
      LET ls_time = cl_replace_str(ls_time, ".", "_")

      LET ls_filename = g_rep_code CLIPPED,"_",ls_date.trim(),"_",ls_time.trim()
   END IF   #171206-00020#1

   RETURN ls_filename
END FUNCTION

################################################################################
# Descriptions...: 依預設值進行排序,優先序為1.用戶(gzgd005) 2.角色(gzgd004) 3.預設樣板(gzgd008)
# Memo...........:
# Usage..........: CALL cl_gr_b_fill_sort()
#                  
# Input parameter: void
# Return code....: none
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_b_fill_sort()
   DEFINE l_i,l_j,l_max  LIKE type_t.num10
   DEFINE l_gzgd_cnt     LIKE type_t.num10
   DEFINE l_gzgd_tmp     DYNAMIC ARRAY OF  type_g_gzgd
   DEFINE l_gzgd_t       type_g_gzgd

   IF g_gzgd.getlength() = 0 THEN RETURN END IF

   #先跑一圈，賦予每一行的樣板排序碼
   LET l_max = 0 
   FOR l_i = 1 TO g_gzgd.getlength()
      IF cl_null(g_gzgd[l_i].gzgd002) THEN CONTINUE FOR END IF
      LET l_gzgd_tmp[l_i].sort_num = 0

      #用戶
      IF g_gzgd[l_i].gzgd005 = g_user THEN
         IF g_gzgd[l_i].gzgd008 = "Y" THEN 
            LET l_gzgd_tmp[l_i].sort_num = l_gzgd_tmp[l_i].sort_num + 1
         ELSE
            LET l_gzgd_tmp[l_i].sort_num = l_gzgd_tmp[l_i].sort_num + 2
         END IF           
      END IF

      #角色
#      IF g_gzgd[l_i].gzgd004 = g_site THEN
      IF g_roles.getIndexOf(g_gzgd[l_i].gzgd004 CLIPPED,1) > 0 THEN
         IF g_gzgd[l_i].gzgd008 = "Y" THEN 
            LET l_gzgd_tmp[l_i].sort_num = l_gzgd_tmp[l_i].sort_num + 4 
         ELSE
            LET l_gzgd_tmp[l_i].sort_num = l_gzgd_tmp[l_i].sort_num + 8
         END IF          
      END IF

      #預設樣版
      IF g_gzgd[l_i].gzgd008 = 'Y' THEN
         IF g_gzgd[l_i].gzgd008 = "Y" THEN 
            LET l_gzgd_tmp[l_i].sort_num =l_gzgd_tmp[l_i].sort_num + 16 
         ELSE
            LET l_gzgd_tmp[l_i].sort_num =l_gzgd_tmp[l_i].sort_num + 32
         END IF           
      END IF

      #IF g_gdw[l_i].gdw06=g_sma.sma124 THEN
      #   LET l_gdw_tmp[l_i].sort_num = l_gdw_tmp[l_i].sort_num + 8 
      #END IF

      #140617 mark 客製已經先過濾
      #IF g_gzgd[l_i].gzgd003='Y' THEN
      #   LET l_gzgd_tmp[l_i].sort_num = l_gzgd_tmp[l_i].sort_num + 16  
      #END IF
      IF l_gzgd_tmp[l_i].sort_num > l_max THEN
         LET l_max = l_gzgd_tmp[l_i].sort_num 
      END IF
   END FOR

   LET l_gzgd_cnt = l_gzgd_tmp.getlength()
   #用選擇排序法
   #FOR l_i = 1 TO l_gzgd_tmp.getlength()
      #FOR l_j = l_i+1 TO l_gzgd_tmp.getlength()
   FOR l_i = 1 TO l_gzgd_cnt
      FOR l_j = l_i+1 TO l_gzgd_cnt        
         IF l_gzgd_tmp[l_j].sort_num > l_gzgd_tmp[l_i].sort_num THEN
            #資料互換
            INITIALIZE l_gzgd_t.* TO NULL
            LET l_gzgd_t.gzgdl002 = g_gzgd[l_i].gzgdl002
            LET l_gzgd_t.gzgd003 = g_gzgd[l_i].gzgd003
            LET l_gzgd_t.gzgd007 = g_gzgd[l_i].gzgd007
            LET l_gzgd_t.gzgd009 = g_gzgd[l_i].gzgd009  #列印簽核
            LET l_gzgd_t.gzgd010 = g_gzgd[l_i].gzgd010  #列印簽核位置
            LET l_gzgd_t.gzgd002 = g_gzgd[l_i].gzgd002 
            LET l_gzgd_t.gzgd000 = g_gzgd[l_i].gzgd000
            LET l_gzgd_t.gzgd005 = g_gzgd[l_i].gzgd005
            LET l_gzgd_t.gzgd004 = g_gzgd[l_i].gzgd004
            LET l_gzgd_t.gzgd008 = g_gzgd[l_i].gzgd008
            #LET l_gzgd_t.gdw06 = gzgd[l_i].gdw06 #FUN-D40043
            LET l_gzgd_t.sort_num = l_gzgd_tmp[l_i].sort_num
                
            LET g_gzgd[l_i].gzgdl002 = g_gzgd[l_j].gzgdl002
            LET g_gzgd[l_i].gzgd003 = g_gzgd[l_j].gzgd003
            LET g_gzgd[l_i].gzgd007 = g_gzgd[l_j].gzgd007
            LET g_gzgd[l_i].gzgd009 = g_gzgd[l_j].gzgd009  #列印簽核
            LET g_gzgd[l_i].gzgd010 = g_gzgd[l_j].gzgd010  #列印簽核位置
            LET g_gzgd[l_i].gzgd002 = g_gzgd[l_j].gzgd002
            LET g_gzgd[l_i].gzgd000 = g_gzgd[l_j].gzgd000
            LET g_gzgd[l_i].gzgd005 = g_gzgd[l_j].gzgd005
            LET g_gzgd[l_i].gzgd004 = g_gzgd[l_j].gzgd004
            LET g_gzgd[l_i].gzgd008 = g_gzgd[l_j].gzgd008
            #LET g_gzgd[l_i].gdw06 = g_gzgd[l_j].gdw06 
            LET l_gzgd_tmp[l_i].sort_num = l_gzgd_tmp[l_j].sort_num
              
            LET g_gzgd[l_j].gzgdl002 = l_gzgd_t.gzgdl002
            LET g_gzgd[l_j].gzgd003 = l_gzgd_t.gzgd003
            LET g_gzgd[l_j].gzgd007 = l_gzgd_t.gzgd007
            LET g_gzgd[l_j].gzgd009 = l_gzgd_t.gzgd009  #列印簽核
            LET g_gzgd[l_j].gzgd010 = l_gzgd_t.gzgd010  #列印簽核位置
            LET g_gzgd[l_j].gzgd002 = l_gzgd_t.gzgd002
            LET g_gzgd[l_j].gzgd000 = l_gzgd_t.gzgd000
            LET g_gzgd[l_j].gzgd005 = l_gzgd_t.gzgd005
            LET g_gzgd[l_j].gzgd004 = l_gzgd_t.gzgd004
            LET g_gzgd[l_j].gzgd008 = l_gzgd_t.gzgd008
            #LET g_gzgd[l_j].gdw06 = l_gzgd_t.gdw06
            LET l_gzgd_tmp[l_j].sort_num = l_gzgd_t.sort_num
         END IF
      END FOR
   END FOR
    
   FOR l_i = 1 TO l_gzgd_tmp.getlength()
      IF l_i=1 THEN
         LET g_gzgd[l_i].pick = 'Y'
      ELSE
         LET g_gzgd[l_i].pick = 'N'
      END IF
   END FOR

END FUNCTION

################################################################################
# Descriptions...: 取得紙張大小
# Memo...........:
# Usage..........: CALL cl_gr_get_paper_size(p_papername)
#                  RETURNING l_width,l_height,l_unit
# Input parameter: p_papername    紙張格式名稱
# Return code....: l_width        紙張寬度
#                : l_height       紙張高度
#                : l_unit         紙張單位
# Date & Author..: 2014 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_paper_size(p_papername)
   DEFINE p_papername  LIKE gzgi_t.gzgi001  #紙張格式名稱
   DEFINE l_width      LIKE gzgi_t.gzgi004  #紙張寬度 
   DEFINE l_height     LIKE gzgi_t.gzgi005  #紙張高度
   DEFINE l_unit       LIKE gzgi_t.gzgi006  #紙張單位 

   SELECT gzgi004,gzgi005,gzgi006 INTO l_width,l_height,l_unit 
     FROM gzgi_t WHERE gzgi001 = p_papername

   RETURN l_width,l_height,l_unit
END FUNCTION

################################################################################
# Descriptions...: 取得4rp樣板檔的紙張格式名稱與紙張方向
# Memo...........:
# Usage..........: CALL cl_gr_get_4rp_paper_size(p_doc_root) 
#                  RETURING l_gzgi001,l_orientation
# Input parameter: p_doc_root     
# Return code....: l_gzgi001      紙張格式名稱
#                : l_orientation  紙張方向
# Date & Author..: 2014 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_get_4rp_paper_size 
################################################################################
PUBLIC FUNCTION cl_gr_get_4rp_paper_size(p_doc_root)
   DEFINE p_doc_root       om.DomNode
   DEFINE l_gzgi001        LIKE gzgi_t.gzgi001
   DEFINE l_orientation    STRING
   #150901-00021#1 mark(s)
#   DEFINE l_node_list      om.NodeList
#   DEFINE l_cur_node       om.DomNode
#   DEFINE l_gzgi004        LIKE gzgi_t.gzgi004
#   DEFINE l_gzgi005        LIKE gzgi_t.gzgi005
#   DEFINE l_gzgi006        LIKE gzgi_t.gzgi006
#   DEFINE l_src_width      STRING
#   DEFINE l_src_length     STRING
#   DEFINE l_pos            LIKE type_t.num10
   #150901-00021#1 mark(e)

   #150901-00021#1 mark(s)   
#   IF p_doc_root IS NOT NULL THEN
#      LET l_node_list = p_doc_root.selectByTagName("report:Settings")
#      IF l_node_list IS NOT NULL AND l_node_list.getLength() >= 1 THEN
#         LET l_cur_node = l_node_list.item(1)
#         LET l_src_width = l_cur_node.getAttribute("RWPageWidth")
#         LET l_src_length = l_cur_node.getAttribute("RWPageLength")
#                    
#         LET l_pos = l_src_width.getIndexOf("width",1)
#         IF l_pos >= 1 THEN
#            LET l_gzgi001 = l_src_width.subString(1,l_pos-1) 
#            LET l_orientation = "P"
#         ELSE
#            LET l_pos = l_src_width.getIndexOf("length",1)
#            IF l_pos >= 1 THEN
#               LET l_gzgi001 = l_src_width.subString(1,l_pos-1) 
#               LET l_orientation = "L"
#            ELSE
#               LET l_pos = l_src_width.getIndexOf("cm",1)
#               IF l_pos >= 1 THEN
#                  LET l_gzgi004 = l_src_width.subString(1,l_pos-1)
#                  LET l_gzgi006 = "C"
#               ELSE
#                  LET l_pos = l_src_width.getIndexOf("in",1)
#                  IF l_pos >= 1 THEN
#                     LET l_gzgi004 = l_src_width.subString(1,l_pos-1)
#                     LET l_gzgi006 = "I"
#                  END IF
#               END IF
#            END IF
#         END IF
#
#         IF l_gzgi006 = "C" THEN
#            LET l_pos = l_src_length.getIndexOf("cm",1)
#            IF l_pos >= 1 THEN
#               LET l_gzgi005 = l_src_length.subString(1,l_pos-1)
#            END IF
#         END IF
#
#         IF l_gzgi006 = "I" THEN
#            LET l_pos = l_src_length.getIndexOf("in",1)
#            IF l_pos >= 1 THEN
#               LET l_gzgi005 = l_src_length.subString(1,l_pos-1)
#            END IF
#         END IF
#
#         IF l_gzgi006 = "C" OR l_gzgi006 = "I" THEN
#            LET l_gzgi001 = NULL
#            SELECT gzgi001,gzgi004,gzgi005 INTO l_gzgi001,l_gzgi004,l_gzgi005
#              FROM gzgi_t WHERE gzgi004 = l_gzgi004 AND gzgi005 = l_gzgi005
#               AND gzgi006 = l_gzgi006
#            IF l_gzgi001 IS NULL THEN
#               SELECT gzgi001,gzgi004,gzgi005 INTO l_gzgi001,l_gzgi004,l_gzgi005
#                 FROM gzgi_t WHERE gzgi004 = l_gzgi005 AND gzgi005 = l_gzgi004
#                  AND gzgi006 = l_gzgi006
#
#               IF l_gzgi001 IS NULL THEN  
#                  LET l_orientation = NULL 
#               ELSE 
#                  LET l_orientation = "L"
#               END IF 
#            ELSE
#               LET l_orientation = "P"
#            END IF
#         ELSE
#            #從4rp抓取紙張大小時需轉換為英文大寫
#            LET l_gzgi001 = UPSHIFT(l_gzgi001)
#         END IF
#      END IF
#   END IF
   #150901-00021#1 mark(e)
   
   CALL cl_rpt_get_4rp_paper_size(p_doc_root) RETURNING l_gzgi001,l_orientation   #150901-00021#1 add
   RETURN l_gzgi001,l_orientation
END FUNCTION

################################################################################
# Descriptions...: T100簽核
# Memo...........:
# Usage..........: CALL cl_gr_tiptop_apr(p_docno)
# Input parameter: p_docno        單號
# Return code....: void
# Date & Author..: 2014/8/18 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_tiptop_apr(p_docno)
   DEFINE p_docno        STRING              #單據編號
   DEFINE l_oojj002      LIKE oojj_t.oojj002 #單據別
   DEFINE l_oojj005      LIKE oojj_t.oojj005 #簽核代號
   DEFINE l_sql          STRING
   DEFINE l_gzgp004      DYNAMIC ARRAY OF LIKE gzgp_t.gzgp004
   DEFINE li_cnt,li_i    INTEGER
   DEFINE l_i1,l_i2,l_i3 INTEGER 
   DEFINE l_success      LIKE type_t.num5    #回傳單號檢查成功否，TRUE/FALSE
   DEFINE l_slip         STRING              #擷取好的單別

   LET g_grSign.sign_width = 2.6
   LET g_grSign.apr_str01 = ""
   LET g_grSign.apr_str02 = ""
   LET g_grSign.apr_str03 = ""
   LET g_grSign.apr_str04 = ""
   LET g_grSign.apr_str05 = ""
   LET g_grSign.apr_str06 = ""
   LET g_grSign.apr_str07 = ""
   LET g_grSign.apr_str08 = ""
   LET g_grSign.apr_str09 = ""
   LET g_grSign.apr_str10 = ""
   LET g_grSign.apr_str11 = ""
   LET g_grSign.apr_str12 = ""
   
   LET l_oojj005 = ""
   LET l_slip = ""

   #經由單號找到單別(l_oojj002)
   #s_aooi200_get_slip(p_docno) RETURNING l_success, l_slip 失敗就不印
#   CALL s_aooi200_get_slip(p_docno) RETURNING l_success, l_slip   #160323-00014#2 mark
   CALL cl_getslip_get_slip(p_docno) RETURNING l_success, l_slip   #160323-00014#2 add
   
   IF l_success THEN
      LET l_oojj002 = l_slip
      
      #取得單別後至oojj_t取得簽核代號
      LET l_sql = "SELECT oojj005 FROM oojj_t WHERE oojjstus = 'Y' ",
                  "AND oojjent = ? AND oojjsite =? AND oojj001 =? AND oojj002 = ? "
                        
      DECLARE cl_gr_oojj005_cs CURSOR FROM l_sql
      
      #ex:(enterprise,site,報表元件代號,單別) = (99,DSCTC,apmr400_g01,102)
      #第一階段,取得符合的ent,site,g_rep_code,單別 ex:(99,DSCTC,apmr400_g01,102)  
      EXECUTE cl_gr_oojj005_cs USING g_enterprise,g_site,g_rep_code,l_oojj002 INTO l_oojj005  
      
      #第二階段,取得符合的ent,default,g_rep_code,單別 ex:(99,default,apmr400_g01,102)
      IF cl_null(l_oojj005) THEN 
         EXECUTE cl_gr_oojj005_cs USING g_enterprise,"default",g_rep_code,l_oojj002 INTO l_oojj005    
      END IF
      
      #第三階段,取得符合的ent,site,g_rep_code,ALL ex:(99,DSCTC,apmr400_g01,ALL)
      IF cl_null(l_oojj005) THEN   
         EXECUTE cl_gr_oojj005_cs USING g_enterprise,g_site,g_rep_code,"ALL" INTO l_oojj005      
      END IF

      #第四階段,取得符合的ent,default,g_rep_code,ALL ex:(99,default,apmr400_g01,ALL)
      IF cl_null(l_oojj005) THEN
         EXECUTE cl_gr_oojj005_cs USING g_enterprise,"default",g_rep_code,"ALL" INTO l_oojj005         
      END IF
      
      #第五階段,取得符合的ent,site,ALL,ALL ex:(99,DSCTC,ALL,ALL)
      IF cl_null(l_oojj005) THEN
         EXECUTE cl_gr_oojj005_cs USING g_enterprise,g_site,"ALL","ALL" INTO l_oojj005      
      END IF
 
      #第六階段,取得符合的ent,default,ALL,ALL ex:(99,default,ALL,ALL)
      IF cl_null(l_oojj005) THEN
         EXECUTE cl_gr_oojj005_cs USING g_enterprise,"default","ALL","ALL" INTO l_oojj005       
      END IF
 
      CLOSE cl_gr_oojj005_cs
      FREE cl_gr_oojj005_cs
      
      IF cl_null(l_oojj005) THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code =  SQLCA.SQLCODE
         LET g_errparam.extend = 'oojj005 is null:'
         LET g_errparam.popup = FALSE
         CALL cl_err()
      END IF
         
      LET l_sql = "SELECT gzgp004 FROM gzgp_t ",
                  " WHERE gzgpstus = 'Y' AND gzgp001 = ? ",
                  "   AND gzgp003 = '",g_dlang,"'",
                  " ORDER BY gzgp002 "
     
      DECLARE cl_gr_apr_cs CURSOR FROM l_sql
         
      LET li_cnt = 1

      CALL l_gzgp004.clear()
      FOREACH cl_gr_apr_cs USING l_oojj005 INTO l_gzgp004[li_cnt]
         IF SQLCA.sqlcode THEN
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = SQLCA.sqlcode
            LET g_errparam.extend = 'cl_gr_apr_cs FOREACH:'
            LET g_errparam.popup = TRUE
            CALL cl_err()

            EXIT FOREACH
         END IF
         LET li_cnt = li_cnt + 1
      END FOREACH
      CLOSE cl_gr_apr_cs
      FREE cl_gr_apr_cs
      CALL l_gzgp004.deleteElement(li_cnt)

      IF NOT cl_null(l_gzgp004[1]) THEN
         LET g_grSign.apr_str01 = l_gzgp004[1] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[2]) THEN
         LET g_grSign.apr_str02 = l_gzgp004[2] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[3]) THEN
         LET g_grSign.apr_str03 = l_gzgp004[3] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[4]) THEN
         LET g_grSign.apr_str04 = l_gzgp004[4] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[5]) THEN
         LET g_grSign.apr_str05 = l_gzgp004[5] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[6]) THEN
         LET g_grSign.apr_str06 = l_gzgp004[6] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[7]) THEN
         LET g_grSign.apr_str07 = l_gzgp004[7] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[8]) THEN
         LET g_grSign.apr_str08 = l_gzgp004[8] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[9]) THEN
         LET g_grSign.apr_str09 = l_gzgp004[9] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[10]) THEN
         LET g_grSign.apr_str10 = l_gzgp004[10] CLIPPED,":"
      END IF

      IF NOT cl_null(l_gzgp004[11]) THEN
         LET g_grSign.apr_str11 = l_gzgp004[11] CLIPPED,":"
      END IF
      
      IF NOT cl_null(l_gzgp004[12]) THEN
         LET g_grSign.apr_str12 = l_gzgp004[12] CLIPPED,":"
      END IF

   ELSE
      #檢查單號不通過
      #LET g_grPageHeader.sign_show = "N"
#      DISPLAY "單號:",p_docno CLIPPED," 檢查不通過,不列印簽核"              #170725-00086#1 mark
      DISPLAY cl_replace_err_msg(cl_getmsg("lib-00437", g_lang),p_docno)   #170725-00086#1 add #單號檢查不通過,不列印簽核
   END IF

END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_gr_splite(p_str)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2014/08/12 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_splite(p_str)
   DEFINE p_str,l_str STRING
   DEFINE l_mail      base.StringTokenizer
   DEFINE l_m         RECORD
                        m1 LIKE type_t.chr80,    #MAIL address
                        m2 LIKE type_t.chr50,    #Name
                        m7 LIKE type_t.chr1,     #附件檔案格式
                        m8 LIKE type_t.chr20     #Account       #150922-00009#1 add
                      END RECORD
   DEFINE l_cnt  LIKE type_t.num5

   LET l_mail = base.StringTokenizer.create(p_str, ":")
   LET l_cnt = 1
   WHILE l_mail.hasMoreTokens()
       LET l_str = l_mail.nextToken()
#       IF NOT cl_null(l_str) THEN      #170524-00050#1 mark
          	 CASE l_cnt
                WHEN 1
                  LET l_m.m1 = l_str
                WHEN 2
                  LET l_m.m2 = l_str
                WHEN 3
                  LET l_m.m7 = l_str
                WHEN 4                  #150922-00009#1 add
                  LET l_m.m8 = l_str    #150922-00009#1 add
             END CASE
             LET l_cnt = l_cnt + 1
#       END IF                          #170524-00050#1 mark
   END WHILE
   RETURN l_m.m1,l_m.m2,l_m.m7,l_m.m8   #150922-00009#1 add m8
END FUNCTION

################################################################################
# Descriptions...: Genero Report 串javamail時等待檔案產生時間
# Memo...........:
# Usage..........: CALL cl_gr_javamail_buffer(p_contex_file,p_rep_code)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2014/8/12 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_javamail_buffer(p_contex_file,p_rep_code)
   DEFINE p_contex_file             STRING  #l_xml_file 
   DEFINE p_rep_code                LIKE gzde_t.gzde001 
   DEFINE l_xml_str                 STRING   #接參數2
   DEFINE l_gzgd011                 LIKE gzgd_t.gzgd011  #閒置時間
   DEFINE l_timeout,l_time_cnt      LIKE type_t.num10 #預設閒置時間, 累計sleep秒數
   DEFINE l_i,i                     LIKE type_t.num10
   DEFINE l_str                     DYNAMIC ARRAY OF STRING  
   DEFINE ls_para                   STRING 
   DEFINE ls_context,l_read_str     STRING 
   DEFINE ls_pid                    STRING 
   DEFINE ls_context_file           STRING 
   DEFINE ls_cmd                    STRING    
   DEFINE lc_chin                   base.Channel        
   DEFINE l_file_sizeA,l_file_sizeB LIKE type_t.num10
   DEFINE ls_tmp                    STRING
   DEFINE l_mail_limit              LIKE type_t.num15_3   #郵件附檔上限 #170725-00086#1
   DEFINE ls_mail_address           STRING                #收件者Mail Address資訊   #190624-00029#1 add
   #IF (NOT cl_user()) THEN
   #   RETURN
   #END IF

   WHENEVER ERROR CALL cl_err_msg_log
#   IF (NOT cl_setup("AZZ")) THEN
#      RETURN
#   END IF

   #LET l_xml_file = ARG_VAL(1)
   #LET g_prog= ARG_VAL(2)
   LET i=1 

   LET ls_cmd = p_contex_file    
   #DISPLAY ls_cmd
   LET lc_chin = base.Channel.create() #create new 物件
   CALL lc_chin.openFile(ls_cmd,"r") #開啟檔案  
      WHILE TRUE   
             LET l_xml_str =l_xml_str,lc_chin.readLine() #整行讀入
              IF lc_chin.isEof() THEN EXIT WHILE END IF     #判斷是否為最後         
       END WHILE
      CALL lc_chin.close()  

   #CALL cl_used(p_rep_code,g_time,1) RETURNING g_time
   #190606-00010#1 mark(s)
#   WHILE (l_xml_str.getLength()>0 )
#      LET  l_str[i]=l_xml_str.subString(1, l_xml_str.getIndexOf("|",1)-1)
#    #  DISPLAY i,"_ l_str:",l_str[i]
#      LET l_xml_str=l_xml_str.subString(l_xml_str.getIndexOf("|",1)+1,l_xml_str.getLength())
#      LET i=i+1 
#   END WHILE 
   #190606-00010#1 mark(e)
   #190606-00010#1 add(s)
   WHILE (l_xml_str.getLength()>0 )
      LET  l_str[i]=l_xml_str.subString(1, l_xml_str.getIndexOf(g_separator_str,1)-1)
      LET l_xml_str=l_xml_str.subString(l_xml_str.getIndexOf(g_separator_str,1)+3,l_xml_str.getLength())
      LET i=i+1
   END WHILE
   #190606-00010#1 add(e)

   LET g_xml.attach=l_str[1]
   LET g_xml.bcc=l_str[2]   
   LET g_xml.body=l_str[3]
   LET g_xml.cc=l_str[4] 
   LET g_xml.file=l_str[5]
   LET g_xml.mailserver=l_str[6]  
   LET g_xml.passwd=l_str[7]
   LET g_xml.recipient=l_str[8]   
   LET g_xml.sender=l_str[9]
   LET g_xml.serverport=l_str[10] 
   LET g_xml.subject=l_str[11]
   LET g_xml.user=l_str[12]
   
   #紀錄收件者的Mail Address，後面需要判斷該收件者是否寄件成功
   #190624-00029#1 add(s)
   LET ls_mail_address = g_xml.recipient
   LET ls_mail_address = ls_mail_address.subString(1,ls_mail_address.getIndexOf(":",1)-1) 
   #190624-00029#1 add(e)
   
   #LET g_xml.sender = "t100.digiwin@yahoo.com.tw" #141002 mark
    
   SELECT DISTINCT gzgd011 INTO l_gzgd011 FROM gzgd_t WHERE gzgd001 = p_rep_code 
                                                       AND gzgdstus = "Y"
   
   LET l_timeout=1200 #預設20mins
#   IF l_gzgd011=0 THEN                         #170725-00086#1 mark
   IF l_gzgd011=0 OR cl_null(l_gzgd011) THEN    #170725-00086#1 add
      LET l_gzgd011=l_timeout #若沒設閒置時間
   ELSE
      LET l_gzgd011=l_gzgd011*60 #換算為秒
   END IF
   LET l_i=l_gzgd011/20
   IF l_i > 10 THEN LET l_i = 10 END IF         #170725-00086#1 add
   LET l_time_cnt=0
   LET l_file_sizeA=0
   LET l_file_sizeB=0
   #SLEEP 3
   LET l_i = 1        #210329-00008#1 每1秒檢查一次
   
   
   WHILE  ( l_time_cnt < l_gzgd011)
      SLEEP 1
      LET l_file_sizeA=os.path.size(g_xml.attach)
      DISPLAY "file_a:",l_file_sizeA
      #SLEEP l_i      #210329-00008#1
      LET l_time_cnt=l_time_cnt+l_i
      LET l_file_sizeB=os.path.size(g_xml.attach)
      DISPLAY "file_b:",l_file_sizeB
      #IF (l_file_sizeA=l_file_sizeB) AND (l_file_sizeA <> 0 and l_file_sizeB <> 0) AND ( os.Path.exists(g_xml.attach)) THEN
      IF (l_file_sizeA=l_file_sizeB)  AND ( os.Path.exists(g_xml.attach)) THEN
        EXIT WHILE 
      END IF 
    
   END WHILE 
   
   #若超時或檔案不存在時，寫一錯誤訊息至log檔
#   IF l_time_cnt>l_gzgd011 AND NOT os.Path.exists(g_xml.attach) THEN
   IF l_time_cnt>l_gzgd011 THEN  #允許沒有附件的狀況
       ERROR "azz1205 ",g_xml.body," user:",g_user  
       DISPLAY "EXIT PROGRAM"     
       RETURN 
   END IF 
   DISPLAY "g_xml.body:",g_xml.body 
   
   #170725-00086#1 add(s)
   LET l_mail_limit = cl_get_para(g_enterprise,"","A-SYS-0068")
   LET l_mail_limit = l_mail_limit * 1024*1024   #換算成Byte
   
   IF l_file_sizeA > l_mail_limit THEN   #超出設定的郵件上限，改提供URL
      LET g_xml.attach = ""
   ELSE
      LET g_attach_desc = g_gr_filename
   END IF
   #170725-00086#1 add(e)
       
   LET lc_chin = base.Channel.create() #create new 物件
   CALL lc_chin.openFile(g_xml.body,"r") #開啟檔案  
      WHILE TRUE   
             LET l_read_str =l_read_str,lc_chin.readLine() #整行讀入
              IF lc_chin.isEof() THEN EXIT WHILE END IF     #判斷是否為最後         
       END WHILE
      CALL lc_chin.close()  
  
   IF NOT os.Path.exists(g_xml.attach) THEN
#     LET l_read_str = cl_replace_str(l_read_str, "${result}", cl_getmsg("lib-216",g_lang))    #170725-00086#1 mark
     LET l_read_str = cl_replace_str(l_read_str, "${result}", cl_getmsg("lib-00073",g_lang))   #170725-00086#1 add #無報表資料產生
   ELSE
#       LET l_read_str = cl_replace_str(l_read_str, "${result}", cl_getmsg("lib-284",g_lang))    #170725-00086#1 mark
       LET l_read_str = cl_replace_str(l_read_str, "${result}", cl_getmsg("adz-00217",g_lang))   #170725-00086#1 add #執行成功
   END IF

   LET l_read_str = cl_str_replace(l_read_str,"g_attach_desc",g_attach_desc)   #170725-00086#1 add             

   LET ls_cmd = "echo '", l_read_str,"' > ",g_xml.body
   RUN ls_cmd    

   LET ls_cmd = "chmod 777 ", g_xml.attach
   RUN ls_cmd

   CALL cl_jmail() RETURNING ls_tmp
   LET g_err_msg = "cl_jmail: ",ls_tmp   #160912-00001#1-2 add
   CALL cl_gr_logfile()                  #160912-00001#1-2 add
       
   #170914-00030#1 add(s)
   IF g_bgjob <> 'Y' THEN
      IF ls_tmp.getIndexOf("successfully",1) THEN
         #190624-00029#1 add(s)
         IF cl_null(g_send_info) THEN
            LET g_send_info = cl_getmsg_parm("lib-00439",g_lang,ls_mail_address)
         ELSE 
            LET g_send_info = g_send_info,"\n",cl_getmsg_parm("lib-00439",g_lang,ls_mail_address)
         END IF
         #190624-00029#1 add(e)
         #190624-00029#1 mark(s)
#         INITIALIZE g_errparam TO NULL
#         LET g_errparam.code = 'lib-00439'
#         LET g_errparam.extend = ''
#         LET g_errparam.popup = TRUE
#         CALL cl_err()   
         #190624-00029#1 mark(s)         
      ELSE
         #190624-00029#1 add(s)
         IF cl_null(g_send_info) THEN
            LET g_send_info = cl_getmsg_parm("lib-00440",g_lang,ls_mail_address)
         ELSE
            LET g_send_info = g_send_info,"\n",cl_getmsg_parm("lib-00440",g_lang,ls_mail_address)
         END IF
         #190624-00029#1 add(e)
         #190624-00029#1 mark(s)
#         INITIALIZE g_errparam TO NULL
#         LET g_errparam.code = 'lib-00440'
#         LET g_errparam.extend = ''
#         LET g_errparam.popup = TRUE
#         CALL cl_err()         
         #190624-00029#1 mark(e)
      END IF
   END IF
   #170914-00030#1 add(e)
   
   DISPLAY "OK"
#   LET g_xml.attach = NULL
   INITIALIZE g_xml TO NULL       
      
   IF os.Path.delete(g_xml.body CLIPPED) THEN
   END IF

   #CALL cl_used(p_rep_code,g_time,2) RETURNING g_time
   RETURN
END FUNCTION

################################################################################
# Descriptions...: 查詢指定語言別下的報表元件名稱
# Memo...........:
# Usage..........: CALL cl_gr_get_progname(p_gzdel001,p_gzdel002,p_type) 
#                  
# Input parameter: p_gzdel001     規格編號
#                : p_gzdel002     語言別
#                : p_type        
# Return code....: r_text         報表元件名稱
# Date & Author..: 2014/08/21 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_progname(p_gzdel001,p_gzdel002,p_type)
   DEFINE p_gzdel001  LIKE gzdel_t.gzdel001
   DEFINE p_gzdel002  LIKE gzdel_t.gzdel002
   DEFINE p_type      LIKE type_t.num5   
   DEFINE r_gzdel003  LIKE gzdel_t.gzdel003
   DEFINE r_text      STRING
   DEFINE l_sql       STRING

   WHENEVER ERROR CALL cl_err_msg_log

   LET p_gzdel001 = p_gzdel001 CLIPPED
   LET p_gzdel002 = p_gzdel002 CLIPPED

   #若傳入的語言別是空的 (未傳入),取用g_lang
   IF p_gzdel002 IS NULL THEN
      LET p_gzdel002 = g_lang
   END IF
   IF p_type <> 2 THEN
      LET p_type = 1
   END IF

   #取得名稱
   LET l_sql = "SELECT gzdel003 FROM gzdel_t WHERE gzdel001 = ? AND gzdel002 = ? "
   PREPARE cl_gr_gzdel_pre FROM l_sql
   EXECUTE cl_gr_gzdel_pre INTO r_gzdel003 USING p_gzdel001,p_gzdel002

   IF p_type = "1" THEN
      RETURN r_gzdel003
   ELSE 
      #LET r_text = r_gzdel003 CLIPPED, " (" , p_gzdel001 CLIPPED, ")"   #161227-00007#1 mark
      LET r_text = r_gzdel003 CLIPPED, " (" , g_prog CLIPPED, ")"    #161227-00007#1 add
      RETURN r_text.trim()
   END IF
END FUNCTION

################################################################################
# Descriptions...: 清空title變數
# Memo...........:
# Usage..........: CALL cl_gr_title_clear()
# Input parameter: none
# Return code....: void
# Date & Author..: 2014/9/25 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_title_clear()

   #LET g_rep_code_desc = ""   #161227-00007#1 mark 移到報表元件清空
   INITIALIZE g_grPageFooter.* TO NULL   #150611-00032#3 add
   
   LET g_grPageHeader.title0101 = ""
   LET g_grPageHeader.title0102 = ""
   LET g_grPageHeader.title0103 = ""
   LET g_grPageHeader.title0104 = ""

   LET g_grPageHeader.title0201 = ""
   LET g_grPageHeader.title0202 = ""
   LET g_grPageHeader.title0203 = ""
   LET g_grPageHeader.title0204 = ""

   LET g_grPageHeader.title0301 = ""
   LET g_grPageHeader.title0302 = ""
   LET g_grPageHeader.title0303 = ""
   LET g_grPageHeader.title0304 = ""

   LET g_grPageHeader.title0401 = ""
   LET g_grPageHeader.title0402 = ""
   LET g_grPageHeader.title0403 = ""
   LET g_grPageHeader.title0404 = ""

   LET g_grPageHeader.title0501 = ""
   LET g_grPageHeader.title0502 = ""
   LET g_grPageHeader.title0503 = ""
   LET g_grPageHeader.title0504 = ""


   LET g_grPageHeader.title0601 = ""
   LET g_grPageHeader.title0602 = ""
   LET g_grPageHeader.title0603 = ""
   LET g_grPageHeader.title0604 = ""

   LET g_grPageHeader.title0701 = ""
   LET g_grPageHeader.title0702 = ""
   LET g_grPageHeader.title0703 = ""
   LET g_grPageHeader.title0704 = ""

   LET g_grPageHeader.title0801 = ""
   LET g_grPageHeader.title0802 = ""
   LET g_grPageHeader.title0803 = ""
   LET g_grPageHeader.title0804 = ""
   
   LET g_grPageHeader.title0901 = ""
   LET g_grPageHeader.title0902 = ""
   LET g_grPageHeader.title0903 = ""
   LET g_grPageHeader.title0904 = ""

   LET g_grPageHeader.title1001 = ""
   LET g_grPageHeader.title1002 = ""
   LET g_grPageHeader.title1003 = ""
   LET g_grPageHeader.title1004 = ""
   
END FUNCTION

################################################################################
# Descriptions...: 取得抬頭資訊
# Memo...........:
# Usage..........: CALL cl_gr_title(p_gzgk006)
#                  RETURNING r_title
# Input parameter: p_gzgk006      传入参数变量说明1
# Return code....: r_title        回传参数变量说明1
# Date & Author..: 2014/09/25 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_title(p_gzgk006)
   #DEFINE p_gzgk006     LIKE gzgk_t.gzgk006   #顯示內容
   DEFINE p_gzgk006     LIKE type_t.chr10      #顯示內容
   DEFINE r_title       STRING                 #報表名稱  
   DEFINE l_str         STRING
   DEFINE l_success     LIKE type_t.num5       #成功與否
   DEFINE l_ooefl004    LIKE ooefl_t.ooefl004  #組織名稱
   DEFINE l_gzcbl004    LIKE gzcbl_t.gzcbl004  #說明
   #180123-00036#1 add(s)
   DEFINE l_tw_title    STRING                 #中越資料
   DEFINE l_add_lang    LIKE type_t.chr6       #附加語系
   DEFINE l_add_separator   LIKE type_t.chr6   #附加分隔符號
   #180123-00036#1 add(e)
   DEFINE l_dlang_t     LIKE gzzy_t.gzzy001    #備份g_dlang #180402-00004#1 add
   DEFINE l_ctid        LIKE ooef_t.ooef002    #統一編號 #180904-00016#1 add
   DEFINE l_flag        LIKE type_t.chr1       #紀錄是否顯示條碼flag #190212-00037#4 add

   LET r_title = ""
   #180123-00036#1 add(s)
   LET l_add_separator = "/"
   IF g_vn_flag = "1" THEN
      LET l_add_lang = "zh_TW"   #繁體中越樣板
   ELSE
      LET l_add_lang = "zh_CN"   #簡體中越樣板
   END IF
   #180123-00036#1 add(e)
   #180402-00004#1 add(s)
   IF g_vn_flag <> "N" THEN
      LET l_dlang_t = g_dlang
      LET g_dlang = "vi_VN"
   END IF
   #180402-00004#1 add(e)
   
   CASE
      #1:一般電話, 2:行動電話, 3:傳真, 4:電子郵件, 5:Skype
      #6:SkypeOut, 7:QQ, 8:Facebook, 9:網址, 10:發票電話
      WHEN (p_gzgk006="1" OR p_gzgk006="2" OR p_gzgk006="3" OR p_gzgk006="4" OR p_gzgk006="5" 
            OR p_gzgk006="6" OR p_gzgk006="7" OR p_gzgk006="8" OR p_gzgk006="9" OR p_gzgk006="10")
            
         CALL cl_gr_get_scc(161,p_gzgk006) RETURNING l_gzcbl004
         CALL cl_s_aooi350_get_site_contact(g_site,p_gzgk006) RETURNING l_success,r_title
         LET r_title = l_gzcbl004,":",r_title

      WHEN p_gzgk006="11"   #對內公司抬頭
         IF g_chk_lang_data = 0 THEN                   #161121-00016#1 add
            CALL s_aooi100_get_fullname(g_site,g_dlang,"1") RETURNING l_success,r_title
            IF l_success = 0 THEN LET g_chk_lang_data = 1 END IF #161121-00016#1 add
            #180123-00036#1 add(s)
            IF g_vn_flag <> "N" THEN
               CALL s_aooi100_get_fullname(g_site,l_add_lang,"1") RETURNING l_success,l_tw_title
               LET r_title = l_tw_title,l_add_separator,r_title
            END IF
            #180123-00036#1 add(e)
         END IF #161121-00016#1 add

      WHEN p_gzgk006="12"   #對外公司抬頭
         IF g_chk_lang_data = 0 THEN                   #161121-00016#1 add      
            CALL s_aooi100_get_fullname(g_site,g_dlang,"2") RETURNING l_success,r_title
            IF l_success = 0 THEN LET g_chk_lang_data = 1 END IF #161121-00016#1 add
            #180123-00036#1 add(s)
            IF g_vn_flag <> "N" THEN
               CALL s_aooi100_get_fullname(g_site,l_add_lang,"2") RETURNING l_success,l_tw_title
               LET r_title = l_tw_title,l_add_separator,r_title
            END IF
            #180123-00036#1 add(e)
         END IF #161121-00016#1 add          
      WHEN p_gzgk006="13"   #公司地址1
         #180123-00036#1 add(s)
         IF g_vn_flag <> "N" THEN
            LET r_title = cl_getmsg("lib-00201",l_add_lang) CLIPPED,l_add_separator,cl_getmsg("lib-00201",g_dlang) CLIPPED,":",g_addr_1
         ELSE
         #180123-00036#1 add(e)
            LET r_title = cl_getmsg("lib-00201",g_dlang) CLIPPED,":",g_addr_1          
         END IF             #180123-00036#1 add
      WHEN p_gzgk006="14"   #公司地址2
         LET r_title = g_addr_2
      WHEN p_gzgk006="15"   #公司地址3
         LET r_title = g_addr_3
      WHEN p_gzgk006="16"   #報表抬頭         
         IF NOT cl_null(g_rep_code_desc) THEN
            LET r_title = g_rep_code_desc
         ELSE
            #LET r_title = cl_gr_get_progname(g_rep_code,g_dlang,1)   #160921-00007#3 mark
            #LET r_title = cl_get_progname(g_prog,g_dlang,1)           #160921-00007#3 add  #161227-00007#1 mark
            LET r_title = cl_gr_get_progname(g_rep_code,g_dlang,1)    #161227-00007#1 add
            #180123-00036#1 add(s)
            IF g_vn_flag <> "N" THEN
               LET l_tw_title = cl_gr_get_progname(g_rep_code,l_add_lang,1)
               LET r_title = l_tw_title,l_add_separator,r_title
            END IF
            #180123-00036#1 add(e)
         END IF
      #150611-00032#3 add(s)
      WHEN p_gzgk006="17"   #報表元件代號
         #LET r_title = "(",g_rep_code CLIPPED,")"                 #160921-00007#3 mark
         LET r_title = "(",g_prog CLIPPED,")"                      #160921-00007#3 add
      WHEN p_gzgk006="18"   #ISO文號
         #180123-00036#1 add(s)
         IF g_vn_flag <> "N" THEN
            LET r_title = cl_getmsg("lib-00253",l_add_lang) CLIPPED,l_add_separator,cl_getmsg("lib-00253",g_dlang) CLIPPED,":",g_rep_ISO
         ELSE
         #180123-00036#1 add(e)
            LET r_title = cl_getmsg("lib-00253",g_dlang) CLIPPED,": ",g_rep_ISO
         END IF         #180123-00036#1 add
      WHEN p_gzgk006="19"   #列印次數
         LET r_title = g_rep_print_cnt
         #180123-00036#1 add(s)
         IF g_vn_flag <> "N" THEN
            LET r_title = cl_getmsg("lib-00254",l_add_lang) CLIPPED,l_add_separator,cl_getmsg("lib-00254",g_dlang) CLIPPED,": ",r_title.trim()         
         ELSE
         #180123-00036#1 add(e)
            LET r_title = cl_getmsg("lib-00254",g_dlang) CLIPPED,": ",r_title.trim()         
         END IF
      WHEN p_gzgk006="22"   #製表者
         #180123-00036#1 add(s)
         IF g_vn_flag <> "N" THEN
            LET r_title = cl_getmsg("lib-00255","zh_TW") CLIPPED,l_add_separator,cl_getmsg("lib-00255",g_dlang) CLIPPED,": ",cl_get_username(g_user)
         ELSE
         #180123-00036#1 add(e)
            LET r_title = cl_getmsg("lib-00255",g_dlang) CLIPPED,": ",cl_get_username(g_user)
         END IF   #180123-00036#1 add
      #150611-00032#3 add(e)
#      #190212-00037#4 mark(s)
#      #161214-00061#1 add(s)
#      WHEN p_gzgk006="24"   #條碼
#         LET r_title = g_code_value
#      #161214-00061#1 add(e)
      #190212-00037#4 mark(e)
      #190212-00037#4 add(s)
      WHEN p_gzgk006 ="24"
         CALL cl_gr_chk_code() RETURNING l_flag
         IF l_flag = "Y" THEN
            LET r_title = g_code_value
         END IF
      #190212-00037#4 (e)
      #180420-00001#4 add(s)
      WHEN p_gzgk006 = "25"
         LET r_title = g_title_def1
      WHEN p_gzgk006 = "26"
         LET r_title = g_title_def2
      WHEN p_gzgk006 = "27"
         LET r_title = g_title_def3
      WHEN p_gzgk006 = "28"
         LET r_title = g_title_def4
      WHEN p_gzgk006 = "29"
         LET r_title = g_title_def5
      #180420-00001#4 add(e)
      #180904-00016#1 add(s)
      WHEN p_gzgk006 = "30"
         SELECT ooef002 INTO l_ctid
           FROM ooef_t
          WHERE ooefent = g_enterprise
            AND ooef001 = g_site
         IF g_vn_flag <> "N" THEN
            LET r_title = cl_getmsg("lib-00450","zh_TW") CLIPPED,l_add_separator,cl_getmsg("lib-00450",g_dlang) CLIPPED,": ",l_ctid
         ELSE
            LET r_title = cl_getmsg("lib-00450",g_dlang) CLIPPED,": ",l_ctid
         END IF
      #180904-00016#1 add(e)
   END CASE
   
   #180402-00004#1 add(s)
   IF g_vn_flag <> "N" THEN
      LET g_dlang = l_dlang_t
   END IF
   #180402-00004#1 add(e)
   
   RETURN r_title
END FUNCTION

################################################################################
# Descriptions...: 取得抬頭SCC說明
# Memo...........:
# Usage..........: CALL cl_gr_get_scc(pc_gzca001,pc_gzcb002)
#                  RETURNING r_gzcbl004
# Input parameter: pc_gzca001     系統分類碼
#                : pc_gzcb002     系統分類碼值
# Return code....: r_gzcbl004     說明
# Date & Author..: 2014/9/25 By 04010
# Modify.........: 
################################################################################
PUBLIC FUNCTION cl_gr_get_scc(pc_gzca001,pc_gzcb002)
   DEFINE pc_gzca001     LIKE gzca_t.gzca001      #系統分類碼
   DEFINE pc_gzcb002     LIKE gzcb_t.gzcb002      #系統分類碼值
   DEFINE r_gzcbl004     LIKE gzcbl_t.gzcbl004    #說明

   LET r_gzcbl004 = "" 
   IF NOT cl_null(pc_gzca001) THEN
       SELECT gzcbl004 INTO r_gzcbl004 FROM gzcbl_t
        WHERE gzcbl001 = pc_gzca001 AND gzcbl002 = pc_gzcb002 AND gzcbl003 = g_dlang
      #190705-00001#1 mark(s)
#      IF SQLCA.SQLCODE THEN
#         INITIALIZE g_errparam TO NULL
#         LET g_errparam.code =  "lib-00068"
#         LET g_errparam.extend = "gzcb_t"
#         LET g_errparam.popup = TRUE
#         CALL cl_err()
#
#         RETURN
#      END IF
      #190705-00001#1 mark(e)
      #190705-00001#1 add(s)
      IF SQLCA.SQLCODE AND g_scc_flag = "N" THEN 
         LET g_scc_flag = "Y"
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code =  "lib-00649"
         LET g_errparam.extend = ""
         LET g_errparam.popup = TRUE
         CALL cl_err()
         RETURN NULL
      END IF
      #190705-00001#1 add(e)
   END IF

   RETURN r_gzcbl004
END FUNCTION

################################################################################
# Descriptions...: 取得公司地址
# Memo...........:
# Usage..........: CALL cl_gr_get_addr(pc_gzgm003)
#                  RETURNING 回传参数
# Input parameter: pc_gzgm003     表頭代號
#                : p_ooek004      列印國家地區
#                : p_ooek005      列印州/省/直轄市
# Return code....: none
# Date & Author..: 2014/9/25 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_addr(pc_gzgm003,p_ooek004,p_ooek005)
   #DEFINE pc_gzgm003     LIKE gzgm_t.gzgm003      #表頭代號   
   DEFINE pc_gzgm003     LIKE type_t.chr10      #表頭代號  
   DEFINE p_ooek004      LIKE ooek_t.ooek004    #列印國家地區
   DEFINE p_ooek005      LIKE ooek_t.ooek005    #列印州/省/直轄市   
   DEFINE l_cnt          LIKE type_t.num5
   DEFINE l_addr         STRING
   DEFINE l_success      LIKE type_t.num5       #成功與否
   
   LET g_addr_1 = ""
   LET g_addr_2 = ""
   LET g_addr_3 = ""
   
#   SELECT COUNT(*) INTO l_cnt FROM gzgk_t WHERE gzgk001 = pc_gzgm003 AND gzgk002 = g_dlang   #170525-00105#1 mark
   SELECT COUNT(1) INTO l_cnt FROM gzgk_t WHERE gzgk001 = pc_gzgm003 AND gzgk002 = g_dlang    #170525-00105#1 add
    AND (gzgk006 = "14" OR gzgk007 = "15" )

   #l_cnt>0, 表頭有配置公司地址2或公司地址3, 地址需抓多行式地址
   IF l_cnt > 0 THEN
      CALL cl_s_aooi350_get_site_addr(g_site,g_dlang,"2",p_ooek004,p_ooek005) RETURNING l_success,l_addr
      
      IF l_addr.getIndexOf("\r\n",1) > 0 THEN
         LET g_addr_1 = l_addr.subString(1,l_addr.getIndexOf("\r\n",1)-1)
         LET l_addr = l_addr.subString(l_addr.getIndexOf("\r\n",1)+2, l_addr.getLength())
      ELSE
         LET g_addr_1 = l_addr
      END IF

      IF l_addr.getIndexOf("\r\n",1) > 0 THEN
         LET g_addr_2 = l_addr.subString(1,l_addr.getIndexOf("\r\n",1)-1)
         LET l_addr = l_addr.subString(l_addr.getIndexOf("\r\n",1)+2, l_addr.getLength())
      ELSE
         LET g_addr_2 = l_addr
      END IF   

      LET g_addr_3 = l_addr
   ELSE
      CALL cl_s_aooi350_get_site_addr(g_site,g_dlang,"1",p_ooek004,p_ooek005) RETURNING l_success,g_addr_1
   END IF

   #IF NOT cl_null(g_addr_1) THEN   #141024 mark
   #   LET g_addr_1 = cl_getmsg("lib-00201",g_lang) CLIPPED,":",g_addr_1   #141024 mark
   #END IF   #141024 mark
   
END FUNCTION

################################################################################
# Descriptions...: 設定g_dlang
# Memo...........:
# Usage..........: CALL cl_gr_set_dlang()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2014/10/1 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_set_dlang()
   DEFINE l_gzzy001 LIKE gzzy_t.gzzy001


   OPEN WINDOW w_cl_gr_s03 WITH FORM cl_ap_formpath("lib","cl_gr_s03")         
      ATTRIBUTES(STYLE="layoutwin")
   CALL cl_ui_init()

   CALL cl_set_combo_lang("gzzy001")
   IF cl_null(g_dlang) THEN
      LET l_gzzy001 = g_lang   #預設為g_lang
      LET g_dlang = g_lang
   ELSE
      LET l_gzzy001 = g_dlang   
   END IF

   DISPLAY l_gzzy001 TO formonly.gzzy001
   WHILE TRUE
      DIALOG ATTRIBUTES(UNBUFFERED)
      
         INPUT l_gzzy001 FROM gzzy001 ATTRIBUTES(WITHOUT DEFAULTS)
            #ON CHANGE gzzy001
            #   LET g_dlang = l_gzzy001
         END INPUT

         #ON ACTION exit
         #   LET INT_FLAG = TRUE
         #   EXIT DIALOG
          
         ON ACTION s_accept
            LET g_dlang = l_gzzy001
#            LET INT_FLAG = TRUE    #151201-00015#1 mark
            LET g_rep_flag = TRUE   #151201-00015#1 add
            EXIT DIALOG            

         ON ACTION close
#            LET INT_FLAG = TRUE    #151201-00015#1 mark
            LET g_rep_flag = TRUE   #151201-00015#1 add
            EXIT DIALOG
            
      END DIALOG
#      IF INT_FLAG THEN             #151201-00015#1 mark
      IF g_rep_flag THEN            #151201-00015#1 add
         #LET INT_FLAG = TRUE       #給TRUE回主程式會跳開 20141003 mark
#         LET INT_FLAG = FALSE      #151201-00015#1 mark
         LET g_rep_flag = FALSE     #151201-00015#1 add
         EXIT WHILE
      END IF   
   END WHILE

   #畫面關閉
   CLOSE WINDOW w_cl_gr_s03

END FUNCTION

################################################################################
# Descriptions...: 取得郵件寄件人資訊
# Memo...........:
# Usage..........: CALL cl_gr_get_mailsender()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2014/10/2 By 04010
# Modify.........: #160113-00004#1 改為PUBLIC
################################################################################
PUBLIC FUNCTION cl_gr_get_mailsender()
   DEFINE l_gzzm       RECORD
            gzzm001    LIKE gzzm_t.gzzm001,
            gzzm002    LIKE gzzm_t.gzzm002,
            gzzm003    LIKE gzzm_t.gzzm003,
            gzzm004    LIKE gzzm_t.gzzm004,
            gzzm005    LIKE gzzm_t.gzzm005,
            gzzm006    LIKE gzzm_t.gzzm006,
            gzzm007    LIKE gzzm_t.gzzm007,
            gzzm009    LIKE gzzm_t.gzzm009,
            gzzm010    LIKE gzzm_t.gzzm010
                   END RECORD   
   DEFINE l_msg        STRING                
   
   #從gzzm_t抓取設定
   #第一階段, 抓g_prog
   
   #判斷是否啟動JavaMail,如無設定,則return
   SELECT gzzm001,gzzm002,gzzm003,gzzm004,gzzm005,gzzm006,gzzm007,gzzm009,gzzm010 INTO l_gzzm.*
     FROM gzzm_t
    WHERE gzzment = g_enterprise AND gzzm001 = g_prog AND gzzmstus = "Y"
   IF SQLCA.SQLCODE THEN
      INITIALIZE l_gzzm.* TO NULL
      SELECT gzzm001,gzzm002,gzzm003,gzzm004,gzzm005,gzzm006,gzzm007,gzzm009,gzzm010 INTO l_gzzm.*
        FROM gzzm_t
       WHERE gzzment = g_enterprise AND gzzm001 = "REPORT" AND gzzmstus = "Y"

      IF SQLCA.SQLCODE THEN
         INITIALIZE l_gzzm.* TO NULL
         SELECT gzzm001,gzzm002,gzzm003,gzzm004,gzzm005,gzzm006,gzzm007,gzzm009,gzzm010 INTO l_gzzm.*
           FROM gzzm_t
          WHERE gzzment = g_enterprise AND gzzm001 = "DEFAULT" AND gzzmstus = "Y"
         IF SQLCA.SQLCODE THEN
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = "lib-00211"
            LET g_errparam.popup = TRUE
            CALL cl_err()
            INITIALIZE l_gzzm.* TO NULL
            #RETURN l_msg
            RETURN
         END IF
      END IF   
   END IF 

   LET g_xml.mailserver = l_gzzm.gzzm003
   LET g_xml.serverport = l_gzzm.gzzm004
   LET g_xml.user = l_gzzm.gzzm005   
   LET g_xml.passwd = l_gzzm.gzzm006
   LET g_xml.sender = l_gzzm.gzzm010

END FUNCTION

################################################################################
# Descriptions...: GR直接送印
# Memo...........:
# Usage..........: CALL cl_gr_direct_print()
#                  
# Input parameter: 
#                : 
# Return code....: 
#                : 
# Date & Author..: 2014/11/15 By 05177
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_direct_print()
   DEFINE i,j,l_i,l_j,l_ac     LIKE type_t.num5
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
   DEFINE l_gzde008            LIKE gzde_t.gzde008           #客製標示   #170316-00056#1 add
   DEFINE l_gzgd012            LIKE gzgd_t.gzgd012           #中越樣板   #180123-00036#1 add

   WHENEVER ERROR CONTINUE
   
   LET l_roles = cl_replace_str(g_roles,",","','")
   
   SELECT gzde008 INTO l_gzde008 FROM gzde_t WHERE gzde001 = g_rep_code   #取得客製碼 #170316-00056#1 add   #180123-00036#1 add
      
#   LET l_sql = "SELECT gzgn001,gzgn002,gzgn003,gzgn004,gzgn005,gzgn006,gzgn007,gzgn008,gzgn009,gzgn010",   #190710-00007#1 mark
               #" FROM gzgn_t WHERE gzgn003 = '",g_prog CLIPPED,"'",                                           #160817-00023#1 mark
               #" FROM gzgn_t WHERE (gzgn003 = '",g_prog CLIPPED,"') OR (gzgn003 ='",g_parentprog CLIPPED,"')",  #160817-00023 add 增加T類判斷   #161013-00029#1 mark
               #190710-00007#1 mark(s)
#               " FROM gzgn_t WHERE ((gzgn003 = '",g_prog CLIPPED,"') OR (gzgn003 ='",g_parentprog CLIPPED,"'))",  #160817-00023 add 增加T類判斷     #161013-00029#1 add             
#               " AND ((gzgn001 = 'default' AND gzgn002 = 'default')",
               #190710-00007#1 mark(e) 
              #" OR gzgn001 = '",g_user CLIPPED,"' OR gzgn002 IN ('",l_roles CLIPPED,"'))",                     #161013-00029#1 mark
              #" AND gzgnstus = 'Y' ORDER BY gzgn001,gzgn002,gzgn004,gzgn005"                         #161013-00029#1 mark
               #161013-00029#1 add -(s)
#               " OR gzgn001 = '",g_user CLIPPED,"' OR gzgn002 IN ('",l_roles CLIPPED,"'))"   #190710-00007#1 mark
   #190710-00007#1 add(s)
   LET l_sql = "SELECT gzgn001,gzgn002,gzgn003,gzgn004,gzgn005,gzgn006,gzgn007,gzgn008,gzgn009,gzgn010",
               "  FROM gzgn_t ",
               " WHERE ((gzgn001 = 'default' AND gzgn002 = 'default')",
               "    OR gzgn001 = '",g_user CLIPPED,"' OR gzgn002 IN ('",l_roles CLIPPED,"'))"
               IF g_bgjob = "T" THEN
                  LET l_sql = l_sql," AND (gzgn003 ='",g_parentprog CLIPPED,"')"
               ELSE
                  LET l_sql = l_sql," AND (gzgn003 = '",g_prog CLIPPED,"')"
               END IF
   #190710-00007#1 add(e)
               IF NOT cl_null(g_template) THEN
                 LET l_sql = l_sql , " AND gzgn004 ='" ,g_template ,"'"    
               #161103-00014#1 add -(s)
               ELSE
                 LET l_sql = l_sql , " AND gzgn004 LIKE '" ,g_rep_code ,"%'"
               #1161103-00014#1 add -(e)                 
               END IF
               #180123-00036#1 add(s)
               IF g_lang <> 'vi_VN' THEN
                  LET l_sql = l_sql, " AND gzgn004 IN (SELECT gzgd002 FROM gzgd_t WHERE gzgd001 = '",g_rep_code,"' ",
                                     "                        AND gzgd003 = '",l_gzde008,"' AND gzgd012 = 'N' AND gzgdstus = 'Y' )"
               END IF
               #180123-00036#1 add(e)
               LET l_sql = l_sql , " AND gzgnstus = 'Y' ORDER BY gzgn001,gzgn002,gzgn004,gzgn005"                  
               #161013-00029#1 add -(e)


   
   DECLARE cl_gr_dp_curs CURSOR FROM l_sql
   
   CALL l_dp_tmp.clear()
   LET l_i = 1
   FOREACH cl_gr_dp_curs INTO l_dp_tmp[l_i].*
      LET l_i = l_i + 1
   END FOREACH
   CLOSE cl_gr_dp_curs
   FREE cl_gr_dp_curs
   
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
   
   #決定直接送印的優先序：g_user > roles > default
   FOR l_i = 1 TO l_dp_tmp.getLength()
      LET l_str = l_dp_tmp[l_i].gzgn002
      LET l_client_ip = cl_getenv_client_ip()
      CASE
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

#      SELECT gzde008 INTO l_gzde008 FROM gzde_t WHERE gzde001 = g_rep_code   #取得客製碼 #170316-00056#1 add   #180123-00036#1 mark

      SELECT gzgd000,gzgd001,gzgd002,gzgd003,gzgd004,gzgd005,gzgd008,gzgd009,gzgd010
      INTO l_dp_tmp[l_i].gzgd000,l_dp_tmp[l_i].gzgd001,l_dp_tmp[l_i].gzgd002,l_dp_tmp[l_i].gzgd003,
           l_dp_tmp[l_i].gzgd004,l_dp_tmp[l_i].gzgd005,l_dp_tmp[l_i].gzgd008,l_dp_tmp[l_i].gzgd009,
           l_dp_tmp[l_i].gzgd010
      FROM gzgd_t
      WHERE gzgd007 = l_dp_tmp[l_i].gzgn004
        AND gzgd003 = l_gzde008   #依報表元件的客製碼取得樣板清單 #170316-00056#1 add
        AND gzgdstus = 'Y'   #180123-00036#1 add
      
      SELECT gzgdl002 INTO l_dp_tmp[l_i].gzgdl002 
      FROM gzgdl_t 
      WHERE gzgdl000 = l_dp_tmp[l_i].gzgd000 AND gzgdl001 = g_dlang
   END FOR

   #若有多筆直接送印設定，取出優先序最大的直接送印設定
   #若相同優先序有多筆直接送印設定，開窗讓使用者選擇樣板，再將直接送印的資料填入全域變數(g_dp)中
   LET l_flag = FALSE
   FOR l_i = 1 TO l_index.getLength()
      FOR l_j = 1 TO l_index[l_i].getLength()
         IF l_index[l_i,l_j] IS NOT NULL THEN
            #將優先序最大的直接送印設定存入區域變數(l_dp)中
            IF l_index[l_i,l_j].getindexof(",", 1) > 0 THEN
               LET l_tok_param = base.StringTokenizer.createExt(l_index[l_i,l_j],",","",TRUE)
               LET i = 1
               LET l_dp_i = 1
               WHILE l_tok_param.hasMoreTokens()
                  LET l_dp_i = l_tok_param.nextToken()
                  #180123-00036#1 add(s)
                  #非越文登入, 不撈取中越樣板的清單
                  IF g_lang <> 'vi_VN' THEN
                     LET l_gzgd012 = 'N'
                     SELECT gzgd012 INTO l_gzgd012 FROM gzgd_t WHERE gzgd000 = l_dp_tmp[l_dp_i].gzgd000
                     IF l_gzgd012 <> 'N' THEN CONTINUE WHILE END IF
                  END IF
                  #180123-00036#1 add(e)
                  LET l_dp[i].* = l_dp_tmp[l_dp_i].*
                  
                  #開窗會使用到的全域變數(g_gzgd)
                  LET g_gzgd[i].pick = 'N'
                  LET g_gzgd[i].gzgdl002 = l_dp[i].gzgdl002
                  LET g_gzgd[i].gzgd003 = l_dp[i].gzgd003
                  LET g_gzgd[i].gzgd007 = l_dp[i].gzgn004
                  LET g_gzgd[i].gzgd002 = l_dp[i].gzgd002
                  LET g_gzgd[i].gzgd000 = l_dp[i].gzgd000
                  LET g_gzgd[i].gzgd005 = l_dp[i].gzgd005
                  LET g_gzgd[i].gzgd004 = l_dp[i].gzgd004
                  LET g_gzgd[i].gzgd009 = l_dp[i].gzgd009
                  LET g_gzgd[i].gzgd010 = l_dp[i].gzgd010
                  LET g_gzgd[i].sort_num = i
                  
                  LET i = i + 1
               END WHILE
            ELSE
               LET l_dp_i = l_index[l_i,l_j]
               LET l_dp[1].* = l_dp_tmp[l_dp_i].*
            END IF
            LET l_flag = TRUE
            EXIT FOR
         END IF
      END FOR
      
      IF l_flag = TRUE THEN
         EXIT FOR
      END IF
   END FOR
   
   #若相同優先序(l_dp)有多筆直接送印設定，開窗讓使用者選擇樣板
   #IF l_dp.getLength() > 1 THEN   #190710-00007#1 mark
   IF l_dp.getLength() > 1 AND cl_null(g_save_flag) THEN   #190710-00007#1 add
      LET l_rootnode = ui.Interface.getRootNode()
      LET l_nodes = l_rootnode.selectByPath("//Window[@name=\"cl_gr_w\"]")
      IF l_nodes.getLength() <= 0 THEN
         OPEN WINDOW cl_gr_w WITH FORM cl_ap_formpath("lib","cl_gr_s01")         
            ATTRIBUTES(STYLE="layoutwin") 
      END IF
      
      CALL cl_ui_init()
      
      CALL cl_set_comp_visible("layout_edit",FALSE)
      CALL cl_set_comp_visible("sampledata",FALSE)
      CALL cl_set_comp_visible("outputformat",FALSE)
      CALL cl_set_comp_visible("set_default",FALSE)
      
      DIALOG ATTRIBUTE(UNBUFFERED)
         DISPLAY ARRAY g_gzgd TO s_gzgd.* ATTRIBUTE(COUNT = g_rec_b)
            BEFORE DISPLAY
               LET l_i = 1
               LET g_gzgd[1].pick = 'Y'                 #初始狀態第一行必勾選
               IF g_gzgd.getLength() != 1 THEN          #若不只一筆樣板檔
                  FOR l_i = 2 TO g_gzgd.getLength()     #其餘筆數均預設不勾選
                     LET g_gzgd[l_i].pick = 'N'
                  END FOR
                  CALL g_gzgd.deleteElement(l_i)        #刪除系統自動產生的最後一筆空行
                  LET l_ac = 1
               END IF
            
            BEFORE ROW
               IF g_gzgd.getLength() != 1 AND l_ac != 0 THEN       #若不只一筆樣板檔
                  IF l_ac != l_dialog.getCurrentRow("s_gzgd") THEN #選擇新筆資料時
                     LET g_gzgd[l_ac].pick = 'N'                   #將上筆取消勾選
                     LET l_ac = l_dialog.getCurrentRow("s_gzgd")   #取得新筆行數
                     LET g_gzgd[l_ac].pick = 'Y'                   #新筆設置為勾選
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
#            LET INT_FLAG = TRUE    #151201-00015#1 mark
            LET g_rep_flag = TRUE   #151201-00015#1 add
            EXIT DIALOG
         
         ON ACTION CLOSE
#            LET INT_FLAG = TRUE    #151201-00015#1 mark
            LET g_rep_flag = TRUE   #151201-00015#1 add
            EXIT DIALOG
      END DIALOG
      
      #將選取到的樣板存入全域變數(g_dp)，供cl_gr_handler()使用
      FOR i = 1 TO g_gzgd.getLength()
         IF g_gzgd[i].pick = 'Y' THEN
            LET l_gzgn004 = g_gzgd[i].gzgd007
         END IF
      END FOR
      
      FOR i = 1 TO l_dp_tmp.getLength()
         IF l_gzgn004.equals(l_dp[i].gzgn004) THEN
            LET g_dp.* = l_dp[i].*
         END IF
      END FOR
   ELSE
      IF l_dp.getLength() > 0 THEN         #180123-00036#1 add
         LET g_dp.* = l_dp[1].*
      END IF                               #180123-00036#1 add
   END IF
   
   #使用者/角色是否有設定直接送印
   LET g_dp_flag = 'N'
   IF l_dp_i >= 1 THEN
      LET g_dp_flag = 'Y'
   END IF
END FUNCTION

################################################################################
# Descriptions...: 取得語言別清單
# Memo...........:
# Usage..........: CALL cl_gr_lang_list()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2014/11/28 By 04010
# Modify.........: #150901-00021#1 改用cl_rpt_lang_list
################################################################################
PUBLIC FUNCTION cl_gr_lang_list()
   DEFINE r_langs      DYNAMIC ARRAY OF LIKE gzzy_t.gzzy001
   
   #150901-00021#1 mark(s)
#   DEFINE li_count     INTEGER
#   DEFINE ls_sql       STRING
   #150901-00021#1 mark(e)

   CALL r_langs.clear()
   
   #150901-00021#1 mark(s)   
#   LET ls_sql = "SELECT gzzy001 FROM gzzy_t WHERE gzzystus = 'Y' ORDER BY gzzy001"
#   DECLARE cl_gr_lang_list_cur CURSOR FROM ls_sql
#   LET li_count = 1
#   FOREACH cl_gr_lang_list_cur INTO r_langs[li_count]
#      IF SQLCA.sqlcode THEN
#         EXIT FOREACH
#      END IF   
#      LET li_count = li_count + 1
#   END FOREACH
#   CLOSE cl_gr_lang_list_cur
#   FREE cl_gr_lang_list_cur
#   
#   CALL r_langs.deleteElement(li_count)
#
####141128先寫死(s)
#   CALL r_langs.clear()
#   LET r_langs[1] = "zh_TW"
#   LET r_langs[2] = "zh_CN"
####141128先寫死(e)
   #150901-00021#1 mark(e)

   CALL cl_rpt_lang_list() RETURNING r_langs   #150901-00021#1 add
   RETURN r_langs
END FUNCTION

################################################################################
# Descriptions...: 判斷該GR報表是否需要留存歷史報表
# Memo...........:
# Usage..........: CALL cl_gr_history_save()
#                  RETURNING l_save
# Input parameter: 
#                : 
# Return code....: l_save   是否需留存歷史報表(Y/N)
#                : 
# Date & Author..: 2014/12/31 By 05177
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_history_save()
   DEFINE l_count       INTEGER   #用來儲存資料庫資料筆數
   DEFINE l_save        STRING    #是否留存資料(Y/N)
   DEFINE l_roles       STRING
   DEFINE l_sql         STRING

   LET l_count = 0
   LET l_save = 'N'
   
   IF cl_get_para(g_enterprise,g_site,'A-SYS-0033') = "N" THEN
      LET g_save_flag = 'N'
      LET l_save = 'N'
      RETURN l_save
   ELSE
#      SELECT count(*) INTO l_count FROM oojo_t   #170725-00086#1 mark
      SELECT COUNT(1) INTO l_count FROM oojo_t    #170725-00086#1 add
      WHERE oojoent = g_enterprise AND (oojosite = g_site OR oojosite = 'ALL') AND oojo001 = g_prog
   
      IF l_count = 0 THEN       #不須留存歷史報表
         LET g_save_flag = 'N'
         LET l_save = 'N'
         RETURN l_save
      ELSE
         #該程式代號於oojo_t有維護資料，需依照全部->權限->使用者代號順序確認是否需留存報表
         #抓取全部類別(oojo003='1')
#         SELECT count(*) INTO l_count FROM oojo_t  #依oojostus確認是否需留存報表   #170725-00086#1 mark
         SELECT COUNT(1) INTO l_count FROM oojo_t  #依oojostus確認是否需留存報表    #170725-00086#1 add
         WHERE oojoent = g_enterprise AND (oojosite = g_site OR oojosite = "ALL") AND oojo001 = g_prog AND oojo003 = '1' AND oojostus = 'Y'
         IF l_count > 0 THEN    #需要留存  
            LET l_save = 'Y'
            RETURN l_save
         END IF         
         
         #依照權限類別抓資料
         LET l_roles = cl_replace_str(g_roles,",","','")
#         LET l_sql = "SELECT count(*) FROM oojo_t",   #170725-00086#1 mark
         LET l_sql = "SELECT COUNT(1) FROM oojo_t",    #170725-00086#1 add
                     " WHERE oojoent = '",g_enterprise CLIPPED,"' AND (oojosite = '",g_site CLIPPED,"' OR oojosite = 'ALL')",
                     " AND oojo001 = '",g_prog CLIPPED,"' AND oojo003 = '2'",
                     " AND oojo002 IN ('",l_roles CLIPPED,"') AND oojostus = 'Y'"
         DECLARE cl_gr_hs_curs CURSOR FROM l_sql
         FOREACH cl_gr_hs_curs INTO l_count
            
         END FOREACH
   
         IF l_count > 0 THEN    #需要留存  
            LET l_save = 'Y'
            RETURN l_save
         END IF
   
         #依照使用者代號抓資料
#         SELECT count(*) INTO l_count FROM oojo_t  #依oojostus確認是否需留存報表   #170725-00086#1 mark
         SELECT COUNT(1) INTO l_count FROM oojo_t  #依oojostus確認是否需留存報表    #170725-00086#1 add
         WHERE oojoent = g_enterprise AND (oojosite = g_site OR oojosite = "ALL") AND oojo001 = g_prog AND oojo003 = '3' AND oojo002 = g_user AND oojostus = 'Y'
         IF l_count > 0 THEN    #需要留存  
            LET l_save = 'Y'
            RETURN l_save
         END IF
         
         #依照'全部->權限->使用者代號'順序確認，資料庫均無資料，代表不須留存
         LET g_save_flag = 'N'
         LET l_save = 'N'
         RETURN l_save
      END IF
   END IF
END FUNCTION

################################################################################
# Descriptions...: 取得歷史報表的留存份數
# Memo...........:
# Usage..........: CALL cl_gr_chk_oojo004()
#                  RETURNING l_oojo004
# Input parameter: 
#                : 
# Return code....: l_oojo004   歷史報表的留存份數
#                : 
# Date & Author..: 2014/12/31 By 05177
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_chk_oojo004()
   DEFINE l_count    INTEGER    #用來儲存資料庫資料筆數
   DEFINE l_oojo004  LIKE oojo_t.oojo004
   DEFINE l_roles    STRING
   DEFINE l_sql      STRING
   
   LET l_count = 0
#   SELECT count(*) INTO l_count FROM oojo_t WHERE oojo001 = g_prog   #170725-00086#1 mark
   SELECT COUNT(1) INTO l_count FROM oojo_t WHERE oojo001 = g_prog    #170725-00086#1 add
#                                              AND oojoent = g_enterrise   #180402-00004#1 add   #180920-00065#1 mark
                                              AND oojoent = g_enterprise   #180920-00065#1 add
   IF l_count = 0 THEN       #不須留存歷史報表
      RETURN NULL
   ELSE
      #該程式代號於oojo_t有維護資料，需依照全部->權限->使用者代號順序確認是否需留存報表
      #抓取全部類別(oojo003='1')
      SELECT oojo004 INTO l_oojo004 FROM oojo_t  #依oojostus確認是否需留存報表
      WHERE oojoent = g_enterprise AND (oojosite = g_site OR oojosite = "ALL") 
        AND oojo001 = g_prog AND oojo003 = '1' AND oojostus = 'Y'
      IF NOT cl_null(l_oojo004) THEN    #需要留存
         RETURN l_oojo004
      END IF
      
      #依照權限類別抓資料
      LET l_roles = cl_replace_str(g_roles,",","','")
      LET l_sql = "SELECT oojo004 FROM oojo_t",
                  " WHERE oojoent = '",g_enterprise CLIPPED,
                  "' AND (oojosite = '",g_site CLIPPED,"' OR oojosite = 'ALL')",
                  " AND oojo001 = '",g_prog CLIPPED,"' AND oojo003 = '2'",
                  " AND oojo002 IN ('",l_roles CLIPPED,"') AND oojostus = 'Y'"
      DECLARE cl_gr_co_curs CURSOR FROM l_sql
      FOREACH cl_gr_co_curs INTO l_oojo004
      END FOREACH
      IF NOT cl_null(l_oojo004) THEN    #需要留存
         RETURN l_oojo004
      END IF

      #依照使用者代號抓資料
      SELECT oojo004 INTO l_oojo004 FROM oojo_t  #依oojostus確認是否需留存報表
      WHERE oojoent = g_enterprise AND (oojosite = g_site OR oojosite = "ALL") 
        AND oojo001 = g_prog AND oojo003 = '3' AND oojo002 = g_user AND oojostus = 'Y'
      IF NOT cl_null(l_oojo004) THEN    #需要留存
         RETURN l_oojo004
      END IF

      #依照'全部->權限->使用者代號'順序確認，資料庫均無資料，代表不須留存
      RETURN NULL
   END IF
   
   RETURN l_oojo004
END FUNCTION

################################################################################
# Descriptions...: 判斷Client端IP是否符合直接送印(azzi320)設定的IP
# Memo...........:
# Usage..........: CALL cl_gr_direct_print_chkip(p_ip,p_client_ip) RETURNING l_sort
#                  
# Input parameter: p_ip          直接送印設定的IP
#                : p_client_ip   Client端IP
# Return code....: l_sort        IP的優先序
# Date & Author..: 2015/02/24 By 05177
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_direct_print_chkip(p_ip,p_client_ip)
   DEFINE p_ip           LIKE type_t.chr20
   DEFINE p_client_ip    LIKE type_t.chr20
   DEFINE l_sort         LIKE type_t.num5
   DEFINE l_flag         LIKE type_t.num5
   DEFINE li_part        LIKE type_t.num5
   DEFINE li             LIKE type_t.num5
   DEFINE l_count        LIKE type_t.num5
   DEFINE lst_pk         base.StringTokenizer
   DEFINE ls_ip          STRING
   DEFINE ld_ip          DYNAMIC ARRAY OF STRING
   DEFINE ld_client_ip   DYNAMIC ARRAY OF STRING
   
   #IP優先序：10.40.40.10 > 10.40.40.* > 10.40.*.* > 10.*.*.* > *.*.*.* = ALL
   #IP優先序預設為6，代表直接送印設定的IP(p_ip)與Client端IP(p_client_ip)比對不符合
   LET l_sort = 6
   
   IF p_ip = "ALL" THEN
      LET l_sort = l_sort - 1
      RETURN l_sort
   END IF
   
   #將直接送印設定的IP拆成陣列
   LET lst_pk = base.StringTokenizer.create(p_ip CLIPPED, ".")
   CALL ld_ip.clear()
   WHILE lst_pk.hasMoreTokens()
      CALL ld_ip.appendElement()
      LET ld_ip[ld_ip.getLength()] = lst_pk.nextToken()
   END WHILE
   
   #將Client端IP拆成陣列
   LET lst_pk = base.StringTokenizer.create(p_client_ip CLIPPED, ".")
   CALL ld_client_ip.clear()
   WHILE lst_pk.hasMoreTokens()
      CALL ld_client_ip.appendElement()
      LET ld_client_ip[ld_client_ip.getLength()] = lst_pk.nextToken()
   END WHILE
   
   #檢查直接送印設定的IP是否符合規範
   IF ld_ip.getLength() <= 3 OR ld_client_ip.getLength() <= 3 THEN
      RETURN l_sort
   END IF
   
   #控卡10.40..40.10寫法
   LET li = 0
   LET ls_ip = p_ip
   WHILE TRUE
      LET l_count = ls_ip.getIndexOf(".",1)
      IF l_count > 0 THEN
         LET li = li + 1
         LET ls_ip = ls_ip.subString(l_count+1,ls_ip.getLength()) CLIPPED
      ELSE
         EXIT WHILE
      END IF
   END WHILE
   
   IF li > 3 THEN
      RETURN l_sort
   END IF
   
   LET l_flag = FALSE
   FOR li = 1 TO ld_ip.getLength()
      IF ld_ip[li] = "*" THEN
         LET l_flag = TRUE
         CONTINUE FOR
      ELSE
         #若IP為10.40.*.10需控卡(l_flag=TRUE)
         IF l_flag == TRUE THEN
            RETURN l_sort
         ELSE
            TRY
               LET li_part = ld_ip[li]
               IF li_part < 0 OR li_part > 255 THEN
                  RETURN l_sort
               END IF
            CATCH
               RETURN l_sort
            END TRY
         END IF
      END IF
   END FOR
   
   #判斷Client端IP是否符合直接送印設定的IP，若符合則需要決定優先序(優先序1~5，6為不符合)
   IF p_ip = p_client_ip THEN
      LET l_sort = 1
   ELSE
      FOR li = 1 TO ld_ip.getLength()
         IF ld_ip[li] = ld_client_ip[li] THEN
            LET l_sort = l_sort - 1
         ELSE
            IF ld_ip[li] = "*" THEN
               LET l_sort = l_sort - 1
            ELSE
               LET l_sort = 6
            END IF
            EXIT FOR
         END IF
      END FOR
   END IF
   
   RETURN l_sort
END FUNCTION

################################################################################
# Descriptions...: 回寫列印次數
# Memo...........:
# Usage..........: CALL cl_gr_print_cnt(p_cnt)
#                  RETURNING 回传参数
# Input parameter: p_cnt          列印次數
#                : 传入参数变量2   传入参数变量说明2
# Return code....: void           回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2015/03/04 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_print_cnt(p_cnt)
   DEFINE p_cnt     LIKE gzgn_t.gzgn007   #列印次數
   DEFINE l_i       LIKE type_t.num5


   IF cl_null(p_cnt) OR p_cnt < 1 THEN LET p_cnt = 1 END IF   #列印次數最少須為1
   
   FOR l_i = 1 TO p_cnt
      PREPARE cl_gr_print_cnt_pr FROM g_print_cnt_sql
      EXECUTE cl_gr_print_cnt_pr 
      
      IF SQLCA.sqlcode THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'lib-00247'
         LET g_errparam.extend = 'g_print_cnt_sql:',g_print_cnt_sql
         LET g_errparam.popup = TRUE
         CALL cl_err()      
      END IF       
   END FOR

END FUNCTION

################################################################################
# Descriptions...: 產生簽核圖檔
# Memo...........:
# Usage..........: CALL cl_gr_mkimg(p_str,p_format)
#                  RETURNING 回传参数
# Input parameter: p_str          簽名
#                : p_format       圖檔格式
# Return code....: l_filepath     圖檔完整路徑
################################################################################
PUBLIC FUNCTION cl_gr_mkimg(p_str,p_format)
   DEFINE p_str      STRING
   DEFINE p_format   STRING
   DEFINE l_filepath STRING
   DEFINE l_fontname STRING
   DEFINE l_fontsize STRING
   DEFINE l_bgcolor  STRING
   DEFINE l_fcolor   STRING
   DEFINE l_cmd      STRING
   DEFINE l_imgsize  STRING
   DEFINE l_filename STRING
   DEFINE l_time     STRING
   DEFINE l_filetype STRING
   DEFINE l_rndnum   LIKE type_t.num5
   DEFINE l_pid      LIKE type_t.num5
   DEFINE l_buf      base.StringBuffer

   #抓取processID
   LET l_pid = FGL_GETPID()
   
   #抓取系統時間到千分之一秒並移除時間分隔符號
   LET l_time = CURRENT HOUR TO FRACTION(3)
   LET l_buf = base.StringBuffer.create()
   CALL l_buf.append(l_time)
   CALL l_buf.replace(":","",0)
   CALL l_buf.replace(".","",0)
   LET l_time = l_buf.toString()
   
   #產生亂數
   CALL util.Math.srand()
   LET l_rndnum = util.Math.rand(32767)
   
   #組圖檔完整路徑
   LET l_filetype = DOWNSHIFT(p_format)
   LET l_filename = l_pid||"_"||l_time||"_"||l_rndnum||"."||l_filetype
   LET l_filepath = os.PATH.join(FGL_GETENV("TEMPDIR"),l_filename)
   
   #讀取系統級報表簽核圖檔相關參數
   CALL cl_get_para(g_enterprise,"","A-SYS-0034") RETURNING l_fontname  #字型
   CALL cl_get_para(g_enterprise,"","A-SYS-0035") RETURNING l_fontsize  #字型大小
   CALL cl_get_para(g_enterprise,"","A-SYS-0036") RETURNING l_imgsize   #圖檔大小
   CALL cl_get_para(g_enterprise,"","A-SYS-0037") RETURNING l_fcolor    #前景色
   CALL cl_get_para(g_enterprise,"","A-SYS-0038") RETURNING l_bgcolor   #背景色

   LET l_cmd = "/usr/bin/convert -density 120 -gravity center -size '",l_imgsize,"'",
               " -pointsize '",l_fontsize,"' -font '",l_fontname,"'",
               " -fill '",l_fcolor,"' -background '",l_bgcolor,"'",
               " label:\"",p_str CLIPPED,"\" ",l_filepath                              

   RUN l_cmd
   RETURN l_filepath
END FUNCTION

################################################################################
# Descriptions...: 取得預設郵件主旨
# Memo...........:
# Usage..........: CALL cl_gr_mail_subject()
#                  RETURNING r_subject
# Input parameter: none
# Return code....: void
# Date & Author..: 2015/8/6 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_mail_subject()
   DEFINE r_subject      STRING

   #170426-00038#1 add(s)
   IF NOT cl_null(g_xml.subject) OR  NOT cl_null(g_schedule_d[1].gzpe005) THEN
      IF cl_null(g_xml.subject) THEN
         LET r_subject = g_schedule_d[1].gzpe005
      ELSE
         LET r_subject = g_xml.subject
      END IF   
   ELSE   
   #170426-00038#1 add(e)
      LET r_subject = #[T100系統通知]
                      #"\[T100",cl_getmsg("lib-00112",g_lang) CLIPPED,"\]",     #161227-00007#1 mark
                      "\[T100",cl_getmsg("lib-00112",g_lang) CLIPPED,"\]"       #161227-00007#1 add
                      #報表名稱：銷售金額統計表
                      #cl_getmsg("lib-00113",g_lang) CLIPPED," : ",cl_gr_get_progname(g_rep_code,g_lang,1) CLIPPED,"，",   #160921-00007#3 mark
                      #cl_getmsg("lib-00113",g_lang) CLIPPED," : ",cl_get_progname(g_prog,g_lang,2) CLIPPED,"，",           #160921-00007#3 add #161227-00007#1 mark
                      #161227-00007#1 add -(s)
                      IF NOT cl_null(g_rep_code_desc) THEN
                        LET r_subject = r_subject , g_rep_code_desc CLIPPED,"，" 
                      ELSE                     
                        LET r_subject = r_subject, cl_get_progname(g_prog,g_dlang,1) CLIPPED,"，"           #160921-00007#3 add
                      END IF 
                      #據點：鼎捷集團台中分公司
                      LET r_subject = r_subject,cl_getmsg("lib-00114",g_lang) CLIPPED," : ", cl_get_deptname(g_site),"，",
                      #161227-00007#1 add -(e)                   
                      ##據點：鼎捷集團台中分公司                                                     #161227-00007#1 mark
                      #cl_getmsg("lib-00114",g_lang) CLIPPED," : ", cl_get_deptname(g_site),"，",  #161227-00007#1 mark
                      #製表(發送)人員為：tiptop
                      cl_getmsg("lib-00115",g_lang) CLIPPED," : ", cl_get_username(g_user) CLIPPED,"，",
                      #製表(發送)日期為：2015/8/6
                      cl_getmsg("lib-00116",g_lang) CLIPPED," : ", g_today   
   END IF   #170426-00038#1 add
   
   RETURN r_subject 
END FUNCTION

################################################################################
# Descriptions...: 取得報表允許列印方式
# Memo...........: #150803-00008#3 add
# Usage..........: CALL cl_gr_get_gzgo004(p_user,p_role,p_prog)
#                  RETURNING r_gzgo004
# Input parameter: none
# Return code....: r_gzgo004      允許列印方式
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2015/8/7 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_get_gzgo004
################################################################################
PUBLIC FUNCTION cl_gr_get_gzgo004(p_user,p_role,p_prog)
   DEFINE p_user             LIKE gzxa_t.gzxa001
   DEFINE p_role             LIKE gzya_t.gzya001
   DEFINE p_prog             LIKE gzzz_t.gzzz001
   #150901-00021#1 mark(s)
#   DEFINE l_i                LIKE type_t.num5
#   DEFINE l_cnt              LIKE type_t.num5 
#   DEFINE l_sql              STRING   
#   
#   DEFINE l_d_cnt            LIKE type_t.num5   #Word   
#   DEFINE l_x_cnt            LIKE type_t.num5   #Excel   
#   DEFINE l_p_cnt            LIKE type_t.num5   #PDF   
#   DEFINE l_i_cnt            LIKE type_t.num5   #HTML   
#   DEFINE l_c_cnt            LIKE type_t.num5   #CSV
#   DEFINE l_j_cnt            LIKE type_t.num5   #java mail   
#   
#   DEFINE l_gzgo004_r        DYNAMIC ARRAY OF LIKE gzgo_t.gzgo004
#   DEFINE l_gzgo004_str      STRING
#   DEFINE l_roles            STRING
#   DEFINE l_gzgo001          LIKE gzgo_t.gzgo001   
#   DEFINE l_gzgo003          LIKE gzgo_t.gzgo003   
#   DEFINE r_gzgo004          LIKE gzgo_t.gzgo004
   #150901-00021#1 mark(e)

   #150901-00021#1 mark(s)
#   CALL l_gzgo004_r.clear()
#   LET r_gzgo004 = NULL
#   
#   IF NOT cl_null(p_prog) THEN
#      LET l_gzgo003 = p_prog
#   ELSE   
#      LET l_gzgo003 = g_prog
#   END IF   
#
#
#   #先取user/default, 只會有一筆
#   IF NOT cl_null(p_user) THEN
#      LET l_gzgo001 = p_user
#   ELSE
#      LET l_gzgo001 = g_user
#   END IF   
#   
#   LET g_sql = "SELECT gzgo004 FROM gzgo_t WHERE gzgostus = 'Y' AND gzgoent = '",g_enterprise CLIPPED,"' ",
#               " AND gzgo001 = '",l_gzgo001 CLIPPED,"' AND gzgo002 = 'default' AND gzgo003 = '",l_gzgo003 CLIPPED,"' "
#   PREPARE cl_gr_gzgo006_pre1 FROM g_sql
#   
#   LET l_cnt = 0
#   LET l_sql = "SELECT COUNT(1) FROM (",g_sql,") tsed"
#   PREPARE cl_gr_gzgo006_cnt_pre1 FROM l_sql
#   EXECUTE cl_gr_gzgo006_cnt_pre1 INTO l_cnt
#   FREE cl_gr_gzgo006_cnt_pre1   
#   IF l_cnt > 0 THEN
#      EXECUTE cl_gr_gzgo006_pre1 INTO r_gzgo004
#      
#      FREE cl_gr_gzgo006_pre1
#      RETURN r_gzgo004
#   END IF
#    
#   #取default/role, 可能有多筆, 取聯集, 如果其中有一個角色沒設定, 則視為不管控
#   IF NOT cl_null(p_role) THEN
#      LET l_roles = "('",p_role,"')"
#   ELSE
#      LET l_roles = cl_replace_str(g_roles,",","','")
#      LET l_roles = "('",l_roles,"')"
#   END IF   
#
#   LET g_sql = "SELECT gzgo004 FROM gzgo_t WHERE gzgostus = 'Y' AND gzgoent = '",g_enterprise CLIPPED,"' ",
#               " AND gzgo001 = 'default' AND gzgo002 IN ", l_roles CLIPPED," AND gzgo003 = '",l_gzgo003 CLIPPED,"' "
#   PREPARE cl_gr_gzgo006_pre2 FROM g_sql
#   DECLARE cl_gr_gzgo006_cs2 CURSOR FOR cl_gr_gzgo006_pre2
#   
#  
#   LET l_sql = "SELECT COUNT(1) FROM (",g_sql,") tsed"
#   PREPARE cl_gr_gzgo006_cnt_pre2 FROM l_sql
#   EXECUTE cl_gr_gzgo006_cnt_pre2 INTO l_cnt
#   FREE cl_gr_gzgo006_cnt_pre2   
#   IF l_cnt > 0 THEN   
#      LET l_i = 1
#      FOREACH cl_gr_gzgo006_cs2 INTO l_gzgo004_r[l_i]
#         LET l_i = l_i + 1
#      END FOREACH
#      
#      LET l_d_cnt = 0
#      LET l_x_cnt = 0
#      LET l_p_cnt = 0
#      LET l_i_cnt = 0
#      LET l_c_cnt = 0
#      LET l_j_cnt = 0
#      
#      FOR l_i = 1 TO l_gzgo004_r.getLength()
#         IF cl_null(l_gzgo004_r[l_i]) THEN
#         
#            FREE cl_gr_gzgo006_pre2
#            FREE cl_gr_gzgo006_cs2
#            RETURN NULL
#         ELSE
#            LET l_gzgo004_str = l_gzgo004_r[l_i]
#            IF l_gzgo004_str.getIndexOf("D",1)>0 THEN LET l_d_cnt = l_d_cnt +1 END IF
#            IF l_gzgo004_str.getIndexOf("X",1)>0 THEN LET l_x_cnt = l_x_cnt +1 END IF
#            IF l_gzgo004_str.getIndexOf("P",1)>0 THEN LET l_p_cnt = l_p_cnt +1 END IF
#            IF l_gzgo004_str.getIndexOf("I",1)>0 THEN LET l_i_cnt = l_i_cnt +1 END IF
#            IF l_gzgo004_str.getIndexOf("C",1)>0 THEN LET l_c_cnt = l_c_cnt +1 END IF
#            IF l_gzgo004_str.getIndexOf("J",1)>0 THEN LET l_j_cnt = l_j_cnt +1 END IF            
#         END IF
#      END FOR
#      
#      IF l_d_cnt > 0 THEN LET r_gzgo004 = r_gzgo004,"D" END IF
#      IF l_x_cnt > 0 THEN LET r_gzgo004 = r_gzgo004,"X" END IF
#      IF l_p_cnt > 0 THEN LET r_gzgo004 = r_gzgo004,"P" END IF
#      IF l_i_cnt > 0 THEN LET r_gzgo004 = r_gzgo004,"I" END IF
#      IF l_c_cnt > 0 THEN LET r_gzgo004 = r_gzgo004,"C" END IF
#      IF l_j_cnt > 0 THEN LET r_gzgo004 = r_gzgo004,"J" END IF
#      
#      FREE cl_gr_gzgo006_pre2
#      FREE cl_gr_gzgo006_cs2      
#      RETURN r_gzgo004
#   END IF
#
#   #取default/default, 只會有一筆
#   LET g_sql = "SELECT gzgo004 FROM gzgo_t WHERE gzgostus = 'Y' AND gzgoent = '",g_enterprise CLIPPED,"' ",
#               " AND gzgo001 = 'default' AND gzgo002 = 'default' AND gzgo003 = '",l_gzgo003 CLIPPED,"' "
#   PREPARE cl_gr_gzgo006_pre3 FROM g_sql
#   EXECUTE cl_gr_gzgo006_pre3 INTO r_gzgo004
#   FREE cl_gr_gzgo006_pre3
#
#   RETURN r_gzgo004
   #150901-00021#1 mark(e)
   
   RETURN cl_rpt_get_gzgo004(p_user,p_role,p_prog)   #150901-00021#1 add
END FUNCTION

################################################################################
# Descriptions...: 允許列表方式下拉選單
# Memo...........: #150803-00008#3 add
# Usage..........: CALL cl_gr_format_combo(p_gzgo004,p_field,p_type)
#                  RETURNING 
# Input parameter: p_gzgo004      允許印表方式
#                : p_type         
# Return code....: void
# Date & Author..: 2015/8/10 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_format_combo
################################################################################
PUBLIC FUNCTION cl_gr_format_combo(p_gzgo004,p_field,p_type)
   DEFINE p_gzgo004          LIKE gzgo_t.gzgo004
   DEFINE p_field            STRING    
   DEFINE p_type             LIKE type_t.chr1   #G:GR列印視窗, X:XG排程下拉選單, OTHERWISE:GR排程下拉選單
   #150901-00021#1 mark(s)
#   DEFINE ls_values          STRING
#   DEFINE ls_items           STRING
#   DEFINE i                  LIKE type_t.num5
#   DEFINE l_num              LIKE type_t.num5
   #150901-00021#1 mark(e)
   
   #150901-00021#1 mark(s)   
#   LET l_num = 1
#   IF p_type = "G" THEN
#      LET ls_values = "1,"
#      LET ls_items  = "1:",cl_getmsg('lib-00257',g_lang),","
#      LET l_num = l_num + 1
#   END IF   
#   
#   IF NOT cl_null(p_gzgo004) THEN
#      FOR i = 1 TO LENGTH(p_gzgo004)
#         CASE p_gzgo004[i,i]
#            WHEN "D" #WORD
#               CASE p_type
#                  WHEN "G"
#                     LET ls_values = ls_values CLIPPED,"7,"
#                  WHEN "X"
#                     LET ls_values = ls_values CLIPPED,"2,"
#                  WHEN "A"
#                     LET ls_values = ls_values CLIPPED,"D,"                  
#                  OTHERWISE
#                     LET ls_values = ls_values CLIPPED,"6," 
#               END CASE
#               LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('lib-00258',g_lang),","
#               LET l_num = l_num + 1   
#               
#            WHEN "X" #EXCEL(XLS,XLSX)
#               CASE p_type
#                  WHEN "G"
#                     LET ls_values = ls_values CLIPPED,"3,4,5,6,"
#                     LET ls_items  = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00715',g_lang),","
#                     LET l_num = l_num + 1
#                     LET ls_items  = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00716',g_lang),","
#                     LET l_num = l_num + 1
#                     LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00717',g_lang),","
#                     LET l_num = l_num + 1               
#                     LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00718',g_lang),","
#                     LET l_num = l_num + 1                      
#                  WHEN "X"
#                     LET ls_values = ls_values CLIPPED,"3,4,"                 
#                     LET ls_items  = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00715',g_lang),","
#                     LET l_num = l_num + 1
#                     LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00717',g_lang),","
#                     LET l_num = l_num + 1
#                  WHEN "A"
#                     LET ls_values = ls_values CLIPPED,"X,B,E,F,"                 
#                     LET ls_items  = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00715',g_lang),","
#                     LET l_num = l_num + 1
#                     LET ls_items  = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00716',g_lang),","
#                     LET l_num = l_num + 1
#                     LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00717',g_lang),","
#                     LET l_num = l_num + 1               
#                     LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00718',g_lang),","
#                     LET l_num = l_num + 1                      
#                  OTHERWISE
#                     LET ls_values = ls_values CLIPPED,"2,3,4,5,"
#                     LET ls_items  = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00715',g_lang),","
#                     LET l_num = l_num + 1
#                     LET ls_items  = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00716',g_lang),","
#                     LET l_num = l_num + 1                     
#                     LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00717',g_lang),","
#                     LET l_num = l_num + 1               
#                     LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00718',g_lang),","
#                     LET l_num = l_num + 1 
#               END CASE
#                           
#            WHEN "P" #PDF
#               CASE p_type
#                  WHEN "G"
#                     LET ls_values = ls_values CLIPPED,"2,"
#                  WHEN "X"
#                     LET ls_values = ls_values CLIPPED,"1,"                 
#                  WHEN "A"
#                     LET ls_values = ls_values CLIPPED,"P,"                 
#                  OTHERWISE
#                     LET ls_values = ls_values CLIPPED,"1," 
#               END CASE
#               LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00720',g_lang),","
#               LET l_num = l_num + 1   
#            
#            WHEN "I" #HTML, XG沒有
#               CASE p_type
#                  WHEN "G"
#                     LET ls_values = ls_values CLIPPED,"8,"
#                  WHEN "X"
#                     CONTINUE FOR                 
#                  WHEN "A"
#                     LET ls_values = ls_values CLIPPED,"I,"           
#                  OTHERWISE
#                     LET ls_values = ls_values CLIPPED,"7," 
#               END CASE
#               LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00719',g_lang),","
#               LET l_num = l_num + 1   
#            
#            WHEN "C" #CSV, 只有XG有
#               CASE p_type
#                  WHEN "X"
#                     LET ls_values = ls_values CLIPPED,"5," 
#                  WHEN "A"
#                     LET ls_values = ls_values CLIPPED,"C,"
#                  OTHERWISE
#                     CONTINUE FOR 
#               END CASE
#               
#               LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('azz-00721',g_lang),","
#               LET l_num = l_num + 1   
#
#            WHEN "J" #Java Mail
#               IF p_type = "G" THEN
#                  LET ls_values = ls_values CLIPPED,"9," 
#               ELSE
#                  CONTINUE FOR                 
#               END IF
#               LET ls_items = ls_items CLIPPED,l_num USING '&',":",cl_getmsg('lib-00259',g_lang),","
#               LET l_num = l_num + 1   
#
#         END CASE #CASE p_gzgo004[i,i]
#      END FOR
#      LET ls_values = ls_values.subString(1, ls_values.getLength() - 1)
#      LET ls_items = ls_items.subString(1, ls_items.getLength() - 1)      
#   ELSE #cl_null(p_gzgo004) #不設限
#      CASE p_type
#         WHEN "G"
#            LET ls_values = ls_values CLIPPED,"2,3,4,5,6,7,8,9"
#            LET ls_items  = ls_items CLIPPED,"2:",cl_getmsg('azz-00720',g_lang)
#                                        ,",","3:",cl_getmsg('azz-00715',g_lang)
#                                        ,",","4:",cl_getmsg('azz-00716',g_lang)
#                                        ,",","5:",cl_getmsg('azz-00717',g_lang)
#                                        ,",","6:",cl_getmsg('azz-00718',g_lang)
#                                        ,",","7:",cl_getmsg('lib-00258',g_lang)
#                                        ,",","8:",cl_getmsg('azz-00719',g_lang)
#                                        ,",","9:",cl_getmsg('lib-00259',g_lang)
#         WHEN "X"
#            LET ls_values = ls_values CLIPPED,"1,2,3,4,5"                
#            LET ls_items  = ls_items CLIPPED,"1:",cl_getmsg('azz-00720',g_lang)            
#                                        ,",","2:",cl_getmsg('lib-00258',g_lang)
#                                        ,",","3:",cl_getmsg('azz-00715',g_lang)
#                                        ,",","4:",cl_getmsg('azz-00717',g_lang)
#                                        ,",","5:",cl_getmsg('azz-00721',g_lang)
#         WHEN "A"
#            LET ls_values = ls_values CLIPPED,"P,X,B,E,F,D,I,C"
#            LET ls_items  = ls_items CLIPPED,"1:",cl_getmsg('azz-00720',g_lang)
#                                        ,",","2:",cl_getmsg('azz-00715',g_lang)
#                                        ,",","3:",cl_getmsg('azz-00716',g_lang)
#                                        ,",","4:",cl_getmsg('azz-00717',g_lang)
#                                        ,",","5:",cl_getmsg('azz-00718',g_lang)
#                                        ,",","6:",cl_getmsg('lib-00258',g_lang)
#                                        ,",","7:",cl_getmsg('azz-00719',g_lang)
#                                        ,",","8:",cl_getmsg('azz-00721',g_lang)
#         OTHERWISE
#            LET ls_values = ls_values CLIPPED,"1,2,3,4,5,6,7" 
#            LET ls_items  = ls_items CLIPPED,"1:",cl_getmsg('azz-00720',g_lang)
#                                        ,",","2:",cl_getmsg('azz-00715',g_lang)
#                                        ,",","3:",cl_getmsg('azz-00716',g_lang)
#                                        ,",","4:",cl_getmsg('azz-00717',g_lang)
#                                        ,",","5:",cl_getmsg('azz-00718',g_lang)
#                                        ,",","6:",cl_getmsg('lib-00258',g_lang)
#                                        ,",","7:",cl_getmsg('azz-00719',g_lang)
#      END CASE
#   END IF      
#   CALL cl_set_combo_items(p_field, ls_values, ls_items)
   #150901-00021#1 mark(e)
   
   CALL cl_rpt_format_combo(p_gzgo004,p_field,p_type)   #150901-00021#1 add
END FUNCTION

################################################################################
# Descriptions...: 取得郵件預設收件人資訊
# Memo...........: #150803-00008#3 add
# Usage..........: CALL cl_gr_get_default_maillist(p_user,p_role,p_prog)
#                  RETURNING r_subject,r_context,r_maillist
# Input parameter: p_user         用戶
#                : p_role         角色
#                : p_prog         程式代碼
# Return code....: r_subject      信件主旨
#                : r_context      信件內文
#                : r_maillist     收件人清單
# Date & Author..: 2015/8/25 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_get_default_maillist
################################################################################
PUBLIC FUNCTION cl_gr_get_default_maillist(p_user,p_role,p_prog)
   DEFINE p_user             LIKE gzxa_t.gzxa001
   DEFINE p_role             LIKE gzya_t.gzya001
   DEFINE p_prog             LIKE gzzz_t.gzzz001
   #150901-00021#1 (s)
#   DEFINE l_oojq001          LIkE oojq_t.oojq001
#   DEFINE l_oojq002          LIkE oojq_t.oojq002
#   DEFINE l_roles            STRING
#   DEFINE l_sql              STRING
#   DEFINE l_i                LIKE type_t.num5
#   DEFINE l_cnt              LIKE type_t.num5  
   #150901-00021#1 (e)
   DEFINE r_subject          LIKE oojq_t.oojq004
   DEFINE r_context          LIKE oojq_t.oojq005
   #190408-00033#2 mark(s)
#   DEFINE r_maillist        DYNAMIC ARRAY OF RECORD   
#              oojn005      LIKE oojn_t.oojn005,  #1:收件者 2:副本 3:密件副本
#              oojn011      LIKE oojn_t.oojn011,  #聯絡對象類型                 
#              oojn003      LIKE oojn_t.oojn003,  #收件人員工編號   
#              oojn004      LIKE oojn_t.oojn004,  #姓名 
#              oojn006      LIKE oojn_t.oojn006,  #Mail address
#              oojn012      LIKE oojn_t.oojn012,  #進階設定   #190107-00075#6 add              
#              oojn007      LIKE oojn_t.oojn007   #報表mail檔案格式               
#                           END RECORD
   #190408-00033#2 mark(e)
   #190408-00033#2 add(s)
   DEFINE r_maillist        DYNAMIC ARRAY OF RECORD
              oojn005      LIKE gzpe_t.gzpe007,  #1:收件者 2:副本 3:密件副本
              oojn011      LIKE gzpe_t.gzpe012,  #聯絡對象類型
              oojn003      LIKE gzpe_t.gzpe004,  #收件人員工編號 
              oojn004      LIKE gzpe_t.gzpe008,  #姓名 
              oojn006      LIKE gzpe_t.gzpe009,  #Mail address
              oojn012      LIKE gzpe_t.gzpe013,  #進階設定
              oojn007      LIKE gzpe_t.gzpe010   #報表mail檔案格式   
                           END RECORD
   #190408-00033#2 add(e)
                           
   
   #150901-00021#1 mark(s)   
#   CALL r_maillist.clear()
#   LET r_subject = NULL
#   LET r_context = NULL
#   
#   LET l_i = 1
# 
#   IF NOT cl_null(p_prog) THEN
#      LET l_oojq001 = p_prog
#   ELSE   
#      LET l_oojq001 = g_prog
#   END IF 
#   
#   #先取user/default, 只會有一筆
#   IF NOT cl_null(p_user) THEN
#      LET l_oojq002 = p_user
#   ELSE
#      LET l_oojq002 = g_user
#   END IF   
#   
#   LET g_sql = "SELECT oojq004,oojq005 FROM oojq_t WHERE oojqstus = 'Y' AND oojqent = '",g_enterprise CLIPPED,"' ",
#               " AND oojq001 = '",l_oojq001 CLIPPED,"' AND oojq002 = '",l_oojq002 CLIPPED,"' AND oojq003 = 'default' "
#   PREPARE cl_gr_get_mail_pre1 FROM g_sql
#   
#   LET l_cnt = 0
#   LET l_sql = "SELECT COUNT(1) FROM (",g_sql,") tsed"
#   PREPARE cl_gr_get_mail_cnt_pre1 FROM l_sql
#   EXECUTE cl_gr_get_mail_cnt_pre1 INTO l_cnt
#   FREE cl_gr_get_mail_cnt_pre1   
#   IF l_cnt > 0 THEN
#      EXECUTE cl_gr_get_mail_pre1 INTO r_subject,r_context
#
#      LET g_sql = "SELECT oojn005,oojn011,oojn003,oojn004,oojn006,oojn007 FROM oojn_t WHERE oojnent = '",g_enterprise CLIPPED,"' ",
#                  " AND oojn001 = '",l_oojq001 CLIPPED,"' AND oojn009 = '",l_oojq002 CLIPPED,"' AND oojn010 = 'default' "
#      PREPARE cl_gr_get_maillist_pre1 FROM g_sql
#      DECLARE cl_gr_get_maillist_cs1 CURSOR FOR cl_gr_get_maillist_pre1
#      FOREACH cl_gr_get_maillist_cs1 INTO r_maillist[l_i].*
#         LET l_i = l_i + 1
#      END FOREACH      
#
#      CALL r_maillist.deleteElement(l_i)
#      
#      FREE cl_gr_get_mail_pre1
#      RETURN r_subject,r_context,r_maillist
#   END IF   
#   
#   #取default/role, 可能有多筆, 取聯集
#   IF NOT cl_null(p_role) THEN
#      LET l_roles = "('",p_role,"')"
#   ELSE
#      LET l_roles = cl_replace_str(g_roles,",","','")
#      LET l_roles = "('",l_roles,"')"
#   END IF   
#   
#   LET g_sql = "SELECT oojq004,oojq005 FROM oojq_t WHERE oojqstus = 'Y' AND oojqent = '",g_enterprise CLIPPED,"' ",
#               " AND oojq001 = '",l_oojq001 CLIPPED,"' AND oojq002 = 'default' AND oojq003 IN ",l_roles CLIPPED,"  AND ROWNUM = 1 "
#   PREPARE cl_gr_get_mail_pre2 FROM g_sql   
#  
#   LET l_sql = "SELECT COUNT(1) FROM (",g_sql,") tsed"
#   PREPARE cl_gr_get_mail_cnt_pre2 FROM l_sql
#   EXECUTE cl_gr_get_mail_cnt_pre2 INTO l_cnt
#   FREE cl_gr_get_mail_cnt_pre2   
#   IF l_cnt > 0 THEN   
#      EXECUTE cl_gr_get_mail_pre2 INTO r_subject,r_context
#      
#      LET g_sql = "SELECT oojn005,oojn011,oojn003,oojn004,oojn006,oojn007 FROM oojn_t WHERE oojnent = '",g_enterprise CLIPPED,"' ",
#                  " AND oojn001 = '",l_oojq001 CLIPPED,"' AND oojn009 = 'default' AND oojn010 IN ",l_roles CLIPPED," "
#      PREPARE cl_gr_get_maillist_pre2 FROM g_sql
#      DECLARE cl_gr_get_maillist_cs2 CURSOR FOR cl_gr_get_maillist_pre2
#      FOREACH cl_gr_get_maillist_cs2 INTO r_maillist[l_i].*
#         LET l_i = l_i + 1
#      END FOREACH      
#
#      CALL r_maillist.deleteElement(l_i)
#      FREE cl_gr_get_mail_pre2
#      RETURN r_subject,r_context,r_maillist
#   END IF   
#
#   #取default/default, 只會有一筆
#   LET g_sql = "SELECT oojq004,oojq005 FROM oojq_t WHERE oojqstus = 'Y' AND oojqent = '",g_enterprise CLIPPED,"' ",
#               " AND oojq001 = '",l_oojq001 CLIPPED,"' AND oojq002 = 'default' AND oojq003 = 'default' "
#   PREPARE cl_gr_get_mail_pre3 FROM g_sql
#   
#   LET l_cnt = 0
#   LET l_sql = "SELECT COUNT(1) FROM (",g_sql,") tsed"
#   PREPARE cl_gr_get_mail_cnt_pre3 FROM l_sql
#   EXECUTE cl_gr_get_mail_cnt_pre3 INTO l_cnt
#   FREE cl_gr_get_mail_cnt_pre3   
#   IF l_cnt > 0 THEN
#      EXECUTE cl_gr_get_mail_pre3 INTO r_subject,r_context
#
#      LET g_sql = "SELECT oojn005,oojn011,oojn003,oojn004,oojn006,oojn007 FROM oojn_t WHERE oojnent = '",g_enterprise CLIPPED,"' ",
#                  " AND oojn001 = '",l_oojq001 CLIPPED,"' AND oojn009 = 'default' AND oojn010 = 'default' "
#      PREPARE cl_gr_get_maillist_pre3 FROM g_sql
#      DECLARE cl_gr_get_maillist_cs3 CURSOR FOR cl_gr_get_maillist_pre3
#      FOREACH cl_gr_get_maillist_cs3 INTO r_maillist[l_i].*
#         LET l_i = l_i + 1
#      END FOREACH      
#
#      CALL r_maillist.deleteElement(l_i)
#      FREE cl_gr_get_mail_pre3
#      RETURN r_subject,r_context,r_maillist
#   END IF   
   #150901-00021#1 mark(e)   
   
   CALL cl_rpt_get_default_maillist(p_user,p_role,p_prog) RETURNING r_subject,r_context,r_maillist
   
   RETURN r_subject,r_context,r_maillist
END FUNCTION

################################################################################
# Descriptions...: 轉換報表預設郵件格式
# Memo...........: #150803-00008#3 add
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2015/8/27 By 04010
# Modify.........: #150901-00021#1改用cl_rpt_mail_format
################################################################################
PUBLIC FUNCTION cl_gr_mail_format(p_rep_type,p_oojn007)
   DEFINE p_rep_type     LIKE gzde_t.gzde005
   DEFINE p_oojn007      LIKE oojn_t.oojn007

   #150901-00021#1 mark(s)
#   #因為預設郵件格式XG跟GR共用(存DPX字元)，因此需轉換(123)
#   CASE p_rep_type
#      WHEN "X"
#         CASE p_oojn007
#            WHEN "P"
#               LET p_oojn007 = "1"
#            WHEN "D"
#               LET p_oojn007 = "2"
#            WHEN "X"
#               LET p_oojn007 = "3"
#            WHEN "E"
#               LET p_oojn007 = "4"
#            WHEN "C"
#               LET p_oojn007 = "5"    
#         END CASE
#      WHEN "G"
#         CASE p_oojn007
#            WHEN "P"
#               LET p_oojn007 = "1"         
#            WHEN "X"
#               LET p_oojn007 = "2"
#            WHEN "B"
#               LET p_oojn007 = "3"
#            WHEN "E"
#               LET p_oojn007 = "4"
#            WHEN "F"
#               LET p_oojn007 = "5"
#            WHEN "D"
#               LET p_oojn007 = "6"
#            WHEN "I"
#               LET p_oojn007 = "7"             
#         END CASE      
#   END CASE   
#   
#   RETURN p_oojn007
   #150901-00021#1 mark(e)

   RETURN cl_rpt_mail_format(p_rep_type,p_oojn007)   #150901-00021#1 add
END FUNCTION

################################################################################
# Descriptions...: 取得郵件收件者的陣列
# Memo...........:
# Usage..........: CALL cl_gr_get_mail_recipient_array(p_recipient)
#                  RETURNING lr_maillist
# Input parameter: p_recipient   收件者(字串,以;分隔)
# Return code....: lr_maillist   收件者(陣列)
# Date & Author..: 2015/10/01 By 05177
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_mail_recipient_array(p_recipient)
   DEFINE p_recipient         STRING
   DEFINE lr_maillist         DYNAMIC ARRAY OF RECORD
          address             STRING,
          username            STRING,
          account             STRING
          END RECORD
   DEFINE lst_maillist        base.StringTokenizer
   DEFINE lst_mailfield       base.StringTokenizer
   DEFINE ls_mail             STRING
   DEFINE l_i,l_j             LIKE type_t.num5
   DEFINE li_idx              LIKE type_t.num5
   DEFINE l_str               STRING

   LET l_i = 1
   LET lst_maillist = base.StringTokenizer.create(p_recipient, ";")
   WHILE lst_maillist.hasMoreTokens()
      LET ls_mail = lst_maillist.nextToken()
      IF NOT cl_null(ls_mail) THEN
         LET l_j = 1
         LET lst_mailfield = base.StringTokenizer.create(ls_mail, ":")
         WHILE lst_mailfield.hasMoreTokens()
            LET l_str = lst_mailfield.nextToken()
            IF NOT cl_null(l_str) THEN
               CASE l_j
                  WHEN 1
                     LET lr_maillist[l_i].address = l_str
                  WHEN 2
                     LET lr_maillist[l_i].username = l_str
                  WHEN 3
                     LET lr_maillist[l_i].account = l_str
               END CASE
               LET l_j = l_j + 1
            END IF
         END WHILE
         LET l_i = l_i + 1
      END IF
   END WHILE

   RETURN lr_maillist
END FUNCTION

################################################################################
# Descriptions...: 取得加密後的附件
# Memo...........:
# Usage..........: CALL cl_gr_get_enc_file(p_address,p_username,p_attach,p_format)
#                  RETURNING l_attach
# Input parameter: p_address      使用者郵件位址
#                : p_username     使用者名稱
#                : p_attach       欲加密碼的附件
#                : p_format       附件格式
# Return code....: l_attach       加密碼後的附件
# Date & Author..: 2015/10/01 By 05177
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_enc_file(p_address,p_username,p_attach,p_format)
   DEFINE p_address         STRING
   DEFINE p_username        LIKE gzxd_t.gzxd001
   DEFINE p_attach          STRING
   DEFINE p_format          STRING
   DEFINE l_password        STRING
   DEFINE l_idx             LIKE type_t.num5
   DEFINE l_gzxd002         LIKE gzxd_t.gzxd002
   DEFINE l_gzxd002_decode  LIKE gzxd_t.gzxd002
   DEFINE l_attach          STRING
   DEFINE l_cmd             STRING
   DEFINE l_pdfenc          com.dsc.t100.report.enc.Pdfenc
   DEFINE l_msg             STRING
   
   LET l_password = ""
   
   #取出使用者密碼並解碼
   SELECT gzxd002 INTO l_gzxd002 FROM gzxd_t WHERE gzxdent = g_enterprise AND gzxd001 = p_username
   IF NOT cl_null(l_gzxd002) THEN
      CALL cl_hashkey_user_decode(l_gzxd002) RETURNING l_gzxd002_decode
      IF NOT cl_null(l_gzxd002_decode) THEN
         LET l_password = l_gzxd002_decode
      END IF
   END IF
   
   #151021-00013#1 add(s)
   #內部郵件(p_username不為空)若取不到用戶密碼,則以郵件帳號加密
   IF NOT cl_null(p_username) AND cl_null(l_password) THEN
      LET l_password = p_address
      LET l_idx = l_password.getIndexOf("@",1)
      LET l_password = l_password.subString(1,l_idx-1)
   END IF
   #151021-00013#1 add(e)
   
   #外寄郵件一律不加密
   LET l_attach = p_attach
   IF NOT cl_null(l_password) THEN
      #將附件加上密碼
      IF p_format = "1" THEN
         LET l_attach = cl_replace_str(p_attach,".pdf",".enc.pdf")
         LET l_pdfenc = Pdfenc.create()
         CALL l_pdfenc.doPdfEncryption(p_attach,l_attach,l_password) RETURNING l_msg
         DISPLAY "l_msg=",l_msg
      ELSE
         LET l_attach = p_attach CLIPPED,".zip"
         LET l_cmd = "zip -j -q -P '",l_password CLIPPED,"' ",l_attach CLIPPED," ",p_attach CLIPPED
         RUN l_cmd
      END IF
   END IF
   
   RETURN l_attach
END FUNCTION

################################################################################
# Descriptions...: 取得附件加密碼設定
# Memo...........:
# Usage..........: CALL cl_gr_get_gzzm011()
#                  RETURNING l_gzzm011
# Input parameter: 
# Return code....: l_gzzm011       附件加密碼設定
# Date & Author..: 2015/10/01 By 05177
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_gzzm011()
   DEFINE l_gzzm011      LIKE gzzm_t.gzzm011
   
   SELECT gzzm011 INTO l_gzzm011
     FROM gzzm_t
     WHERE gzzment = g_enterprise AND gzzm001 = g_prog AND gzzmstus = "Y"
   IF SQLCA.SQLCODE THEN
      LET l_gzzm011 = "N"
      SELECT gzzm011 INTO l_gzzm011
        FROM gzzm_t
        WHERE gzzment = g_enterprise AND gzzm001 = "REPORT" AND gzzmstus = "Y"

      IF SQLCA.SQLCODE THEN
         LET l_gzzm011 = "N"
         SELECT gzzm011 INTO l_gzzm011
           FROM gzzm_t
           WHERE gzzment = g_enterprise AND gzzm001 = "DEFAULT" AND gzzmstus = "Y"
         IF SQLCA.SQLCODE THEN
            LET l_gzzm011 = "N"
            LET g_errparam.code = "lib-00211"
            LET g_errparam.popup = TRUE
            CALL cl_err()
            LET l_gzzm011 = "N"
            RETURN
         END IF
      END IF   
   END IF
   
   RETURN l_gzzm011
END FUNCTION

################################################################################
# Descriptions...: 紀錄執行log
# Memo...........: 160912-00001#1-2 add
# Usage..........: CALL cl_gr_logfile()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2016/9/22 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_logfile()
   DEFINE l_file  STRING,
          l_s     STRING,
          l_cmd   STRING,
          i       SMALLINT
   DEFINE l_channel       base.Channel       

   LET l_file = "gr-", TODAY USING 'YYYYMMDD', ".log"
   LET l_channel = base.Channel.create()

   CALL l_channel.openFile(l_file,  "a")
   IF STATUS = 0 THEN
      CALL l_channel.setDelimiter("")
      #紀錄傳入的 XML 字串
      LET l_s = "#-------------------------------- (", cl_get_current(), ") --------------------------------#"
      CALL l_channel.write(l_s)
      LET l_s = "Program:",g_prog, " , Repcode:",g_rep_code," , User:",g_user," , Enterprise:",g_enterprise
      LET l_s = l_s," , Account:",g_account
      CALL l_channel.write(l_s)
      LET l_s = "4rp:",g_4rp_path
      CALL l_channel.write(l_s)
      IF NOT cl_null(g_schedule_id) THEN
         LET l_s = "指定時間於背景執行 g_bgjob:",g_bgjob," ,g_schedule_id:",g_schedule_id," ,g_schedule_seq:",g_schedule_seq," ,g_msgparam.rpt_code:",g_msgparam.rpt_code          
         CALL l_channel.write(l_s)
         LET l_s = "g_run_cmd:",g_run_cmd
         CALL l_channel.write(l_s)
         LET l_s = "g_run_result:",g_run_result, " ,g_err_msg:",g_err_msg
         CALL l_channel.write(l_s)
         #170710-00060#1 add(s)
         IF NOT cl_null(g_rs_printer) THEN
            LET l_s = "g_rs_printer:", g_rs_printer
            CALL l_channel.write(l_s)
         END IF
         #170710-00060#1 add(e)
         #190503-00009#1 add(s)
         LET l_s = g_template_info
         CALL l_channel.write(l_s)
         IF NOT cl_null(FGL_GETENV("GUIMODE")) THEN
            LET l_s = "GUIMODE: ",FGL_GETENV("GUIMODE")
            CALL l_channel.write(l_s)
         END IF
         #190503-00009#1 add(e)
      ELSE
         IF g_bgjob = 'Y' THEN
            LET l_s = "立即於背景執行 g_bgjob:",g_bgjob," ,g_schedule_id:",g_schedule_id," ,g_schedule_seq:",g_schedule_seq," ,g_msgparam.rpt_code:",g_msgparam.rpt_code                   
            CALL l_channel.write(l_s)
            LET l_s = "g_run_cmd:",g_run_cmd
            CALL l_channel.write(l_s)
            LET l_s = "g_run_result:",g_run_result, " ,g_err_msg:",g_err_msg
            CALL l_channel.write(l_s)         
            #190107-00075#6 add(s)
            IF g_schedule_d.getLength() > 0 THEN
               LET l_s = "收件者資訊"
               CALL l_channel.write(l_s)
               FOR i = 1 TO g_schedule_d.getLength()
                  LET l_s = "   g_schedule_d[",i,"].gzpe007(形式):",g_schedule_d[i].gzpe007
                  CALL l_channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe004(員工編號):",g_schedule_d[i].gzpe004
                  CALL l_channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe009(Email address):",g_schedule_d[i].gzpe009
                  CALL l_channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe013(進階設定):",g_schedule_d[i].gzpe013
                  CALL l_channel.write(l_s)
                  LET l_s = "   g_schedule_d[",i,"].gzpe010(檔案格式):",g_schedule_d[i].gzpe010
                  CALL l_channel.write(l_s)               
               END FOR
            END IF
            #190107-00075#6 add(e)
            #190503-00009#1 add(s)
            LET l_s = g_template_info
            CALL l_channel.write(l_s)
            IF NOT cl_null(FGL_GETENV("GUIMODE")) THEN
               LET l_s = "GUIMODE: ",FGL_GETENV("GUIMODE")
               CALL l_channel.write(l_s)
            END IF
            #190503-00009#1 add(e)
         ELSE
            LET l_s = "立即於前景執行 "
            CALL l_channel.write(l_s)   
            LET l_s = "g_err_msg:",g_err_msg      #181213-00050#1 add
            #190503-00009#1 add(s)
            LET l_s = g_template_info
            CALL l_channel.write(l_s)
            IF NOT cl_null(FGL_GETENV("GUIMODE")) THEN
               LET l_s = "GUIMODE: ",FGL_GETENV("GUIMODE")
               CALL l_channel.write(l_s)
            END IF
            #190503-00009#1 add(e)
         END IF
      END IF

      #170725-00086#1 add(s)
      IF NOT cl_null(g_mail_resp) THEN
         LET l_s = "郵件下載狀況:",g_mail_resp
         CALL l_channel.write(l_s)
      END IF
      #170725-00086#1 add(e)

      IF g_errlog.getlength() = 0 THEN #有錯誤
         CALL l_channel.write("Successful !")
         LET l_s = "#---------------------------------------------------------------------------------------#"
         CALL l_channel.write(l_s)
      ELSE
         CALL l_channel.write("Error : ")
         FOR i = 1 TO g_errlog.getlength()
            CALL l_channel.write(g_errlog[i])
         END FOR
         LET l_s = "#---------------------------------------------------------------------------------------#"
         CALL l_channel.write(l_s)
      END IF

      CALL l_channel.close()

      IF os.Path.chrwx(l_file CLIPPED,438) THEN END IF 
   END IF

END FUNCTION

################################################################################
# Descriptions...: 壓縮sampledata相關檔案
# Memo...........: #170808-00030#1 add
# Usage..........: CALL cl_gr_sampledata_zip()
#                  RETURNING r_result
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/08/08 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_sampledata_zip()
   DEFINE
      ls_zip_name           STRING,      #壓縮檔名稱
      ls_random_path        STRING,      #UUID
      ls_module_env         STRING,
      ls_dir_pid            STRING,
      li_chdir              INTEGER,     #切換目錄回傳值
      li_mkdir              INTEGER,     #建立目錄回傳值
      lb_result             BOOLEAN
   DEFINE ls_type           STRING   #系統別
   DEFINE ls_module         STRING   #模組別
   DEFINE ls_msg            STRING
   DEFINE ls_src_file       STRING  
   DEFINE ls_dst_file       STRING
   DEFINE ls_sql            STRING
   DEFINE ls_cmd            STRING
   DEFINE ls_module_path    STRING
   DEFINE l_gzgd007         LIKE gzgd_t.gzgd007
   DEFINE ls_dir_zip        STRING   #檔案存放位置, ls_dir_pid下一層
   DEFINE li_stat           LIKE type_t.num5
   DEFINE ls_basename       STRING       #壓縮檔名稱(不含副檔名)
   DEFINE l_date            LIKE type_t.dat

   #取得模組別
   LET ls_module = cl_adz_find_module(g_rep_code, "G")
   IF cl_null(ls_module) THEN
      LET ls_msg = "ERROR:",cl_replace_err_msg(cl_getmsg("adz-00378", g_lang),g_rep_code)
      DISPLAY "E",ls_msg
#      RETURN   #190705-00001#1 mark
      RETURN NULL   #190705-00001#1 add
   END IF

   LET ls_module_env   = ls_module.toUpperCase()   
   LET ls_module_path  = FGL_GETENV(ls_module_env)

   #壓縮檔名稱為 "客戶名稱_區域名稱_報表元件代號.zip"
   LET l_date = cl_get_current()
   LET ls_basename = l_date USING 'YYYYMMDD',"_",FGL_GETENV("CUST") CLIPPED,"_",FGL_GETENV("ZONE"),"_",g_rep_code CLIPPED
   LET ls_zip_name = ls_basename,".zip"

   #以亂數產生資料夾名稱
   CALL security.RandomGenerator.CreateUUIDString() RETURNING ls_random_path
   LET ls_dir_pid = os.Path.join(FGL_GETENV("TEMPDIR"),g_rep_code CLIPPED||"_"||ls_random_path)

   #建立暫存目錄
   CALL os.Path.mkdir(ls_dir_pid) RETURNING li_mkdir

   LET ls_dir_zip = os.Path.join(ls_dir_pid,ls_basename)
   CALL os.Path.mkdir(ls_dir_zip) RETURNING li_mkdir

   #複製rdd, 4rp, sampledata到暫存檔目錄
   LET ls_src_file = os.Path.join(os.Path.join(os.Path.join(ls_module_path,"4rp"),"rdd"),g_rep_code CLIPPED||".rdd")
   LET ls_dst_file = os.Path.join(ls_dir_zip,g_rep_code CLIPPED||".rdd")
   CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result

   LET ls_sql = " SELECT gzgd007 FROM gzgd_t ",
                "  WHERE gzgd001 = '",g_rep_code CLIPPED,"' AND gzgd002 = '",g_gzgd[g_sampledata_ac].gzgd002 CLIPPED,"' ",
                "    AND gzgd003 = '",g_gzgd[g_sampledata_ac].gzgd003 CLIPPED,"' AND gzgd004 = '",g_gzgd[g_sampledata_ac].gzgd004 CLIPPED,"' ",
                "    AND gzgd005 = '",g_gzgd[g_sampledata_ac].gzgd005 CLIPPED,"' "
   DECLARE cl_gr_sampledata_cs CURSOR FROM ls_sql
   FOREACH cl_gr_sampledata_cs INTO l_gzgd007
      LET ls_src_file = os.Path.join(os.Path.join(os.Path.join(ls_module_path,"4rp"),g_rlang),l_gzgd007 CLIPPED||".4rp")
      LET ls_dst_file = os.Path.join(ls_dir_zip,l_gzgd007||".4rp")
      CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result      
      
      #170815-00008#1 add(s)
      #複製圖檔, 限JPG/PNG
      #JPG
      LET ls_src_file = os.Path.join(os.Path.join(os.Path.join(ls_module_path,"4rp"),g_rlang),l_gzgd007 CLIPPED||".jpg")
      LET ls_dst_file = os.Path.join(ls_dir_zip,l_gzgd007||".jpg")
      CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result      
      #PNG
      LET ls_src_file = os.Path.join(os.Path.join(os.Path.join(ls_module_path,"4rp"),g_rlang),l_gzgd007 CLIPPED||".png")
      LET ls_dst_file = os.Path.join(ls_dir_zip,l_gzgd007||".png")
      CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result      
      #170815-00008#1 add(e)     
   END FOREACH

   LET ls_src_file = os.Path.join(os.Path.join(os.Path.join(ls_module_path,"4rp"),"sampledata"),g_gzgd[g_sampledata_ac].gzgd007 CLIPPED||".sampledata")
   LET ls_dst_file = os.Path.join(ls_dir_zip,g_rep_code CLIPPED||".sampledata")
   CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result      


   #壓縮
   CALL os.Path.chdir(ls_dir_pid) RETURNING li_chdir
   LET ls_cmd = "zip -r '",ls_zip_name,"' *"
   RUN ls_cmd

   LET ls_src_file = os.Path.join(ls_dir_pid,ls_zip_name)

   #下載至本機
   LET ls_dst_file = cl_client_browse_dir()   #選擇要下載的路徑
   IF cl_null(ls_dst_file) THEN
      LET li_stat = FALSE
   ELSE
      LET ls_dst_file = os.Path.join(ls_dst_file,ls_zip_name)
      LET li_stat = cl_client_download_file(ls_src_file,ls_dst_file)   #下載檔案
      IF li_stat THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'azz-00359'
         LET g_errparam.extend = 'azz-00359'
         LET g_errparam.popup = TRUE
         CALL cl_err()
      END IF
   END IF
   
   #刪除壓縮檔
   LET ls_cmd = "rm -r ",ls_dir_pid
   RUN ls_cmd

   RETURN lb_result
END FUNCTION

################################################################################
# Descriptions...: 依聯絡人代碼取得電子郵件
# Memo...........: 170929-00002#1 add
# Usage..........: CALL cl_gr_get_cont_ppl_email(p_cont_type,p_id)
#                  RETURNING r_email
# Input parameter: p_cont_type    聯絡對象類型
#                                 0.鼎新顧問團隊 1.組織 2.員工 3.客戶/廠商
#                                 4.交易對象聯絡人 7.銀行 8.庫位 12.站點
#                : p_id           聯絡對象代碼
# Return code....: r_email        電子郵件
# Date & Author..: 2017/10/5 By 04010
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_cont_ppl_email(p_cont_type,p_id)
   DEFINE p_cont_type    LIKE oofa_t.oofa002   #聯絡對象類型
   DEFINE p_id           LIKE oofa_t.oofa003   #聯絡對象代碼
   DEFINE r_email        LIKE oofc_t.oofc012   #電子郵件
   DEFINE l_oofa001      LIKE oofa_t.oofa001   #聯絡對象識別碼
   DEFINE l_success      LIKE type_t.num5
   DEFINE l_gzou003      LIKE gzou_t.gzou003   #企業編號主資料庫
   DEFINE l_dbs_str      STRING
   DEFINE l_sql          STRING


   WHENEVER ERROR CONTINUE

   #變數初始化
   LET r_email = NULL

   #傳入連絡對象類型為空
   IF cl_null(p_cont_type) THEN
      LET l_success = FALSE
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'aoo-00647'
      LET g_errparam.popup = FALSE
      CALL cl_err()
      RETURN r_email
   END IF

   #傳入連絡對象代碼為空
   IF cl_null(p_id) THEN
      LET l_success = FALSE
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'sub-00106'
      LET g_errparam.popup = FALSE
      CALL cl_err()
      RETURN r_email
   END IF

   #抓取企業編號主資料庫,SQL需要組上l_gzou003
   #例如:AZZ/ADZ模組的作業執行時,g_dbs='ds',則此時要抓聯絡對象資料會抓不到,需要改抓dsdemo的資料
   LET l_gzou003 = NULL
   LET l_dbs_str = NULL
   SELECT gzou003 INTO l_gzou003 FROM gzou_t WHERE gzou001 = g_enterprise
   LET l_dbs_str = l_gzou003,"."

   #抓取聯絡對象識別碼
   LET l_oofa001 = ''

   CASE p_cont_type
      WHEN "2"   #2.員工aooi130
         LET l_sql = "SELECT ooag002 FROM ",l_dbs_str CLIPPED,"ooag_t",
                    " WHERE ooagent = ",g_enterprise,
                    "   AND ooag001 = '",p_id CLIPPED,"'",
                    "   ORDER BY ooag001 DESC"
      WHEN "3"   #3.客戶/廠商apmm100/apmm800
         LET l_sql = "SELECT pmaa027 FROM ",l_dbs_str CLIPPED,"pmaa_t",
                    " WHERE pmaaent = ",g_enterprise,
                    "   AND pmaa001 = '",p_id CLIPPED,"'",
                    "   ORDER BY pmaa001 DESC"
      WHEN "4"   #4.交易對象聯絡人apmm100/apmi005
         LET l_sql = "SELECT pmaj002 FROM ",l_dbs_str CLIPPED,"pmaj_t",
                    " WHERE pmajent = ",g_enterprise,
                    "   AND pmaj001 = '",p_id CLIPPED,"'",
                    "   ORDER BY pmaj001 DESC"
      OTHERWISE   #aooi350
         LET l_sql = "SELECT oofa001 FROM ",l_dbs_str CLIPPED,"oofa_t",
                    " WHERE oofaent = ",g_enterprise,
                    "   AND oofa002 = '",p_cont_type CLIPPED,"'",
                    "   AND oofa003 = '",p_id CLIPPED,"'",
                    "   ORDER BY oofa001 DESC"
   END CASE
   
   DISPLAY "l_sql:",l_sql
   DECLARE cl_gr_get_cont_ppl_email_cs CURSOR FROM l_sql


   FOREACH cl_gr_get_cont_ppl_email_cs INTO l_oofa001
      EXIT FOREACH
   END FOREACH

   CLOSE cl_gr_get_cont_ppl_email_cs
   FREE cl_gr_get_cont_ppl_email_cs

   #return3代表開窗選擇的聯絡人識別碼
   IF NOT cl_null(g_qryparam.return3) THEN
      LET l_oofa001 = g_qryparam.return3
   END IF
   
   IF cl_null(l_oofa001) THEN
      RETURN r_email
   END IF

   CALL cl_s_aooi350_get_contact(l_oofa001,'4',l_dbs_str)
        RETURNING l_success,r_email

   RETURN r_email

END FUNCTION

################################################################################
# Descriptions...: 取得主機端系統GRE版本
# Memo...........: #171211-00055#1 add
# Usage..........: CALL cl_gr_get_gre_version()
#                  RETURNING ls_version
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: ls_version     GRE版本
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/12/12 By Cynthia
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_get_gre_version()
    DEFINE li_idx     INTEGER
    DEFINE lch_cmd    base.Channel
    DEFINE ls_buf     STRING,
           ls_version STRING,
           ls_str     STRING,
           l_path     STRING,           
           l_key      STRING
    DEFINE l_str      INTEGER,
           l_end      INTEGER

    WHENEVER ERROR CALL cl_err_msg_log

    #---------------------------------------------------------------------------
    # 利用 'greportwriter -V' 指令擷取 GRE 版本
    #---------------------------------------------------------------------------
    INITIALIZE ls_version TO NULL
    LET lch_cmd = base.Channel.create()
    LET l_path = "$GREDIR/bin/greportwriter"
    LET l_key = "greportwriter"
    CALL lch_cmd.openPipe(l_path || " -V", "r")
    WHILE lch_cmd.read(ls_buf)
        LET ls_str = ls_buf
        LET ls_buf = ls_buf.toLowerCase()
        IF ( li_idx := ls_buf.getIndexOf(l_key, 1) ) THEN
                 LET l_str = li_idx + l_key.getLength()
                 LET l_end = ls_buf.getIndexOf(" ", l_str + 1) - 1
           LET ls_version = ls_str.subString(l_str, l_end)
           LET ls_version = ls_version.trim()
           EXIT WHILE
        END IF
    END WHILE
    CALL lch_cmd.close()
    
    RETURN ls_version

END FUNCTION

################################################################################
# Descriptions...: 取得用戶是否有設保留列印視窗值
# Memo...........:
# Usage..........: CALL cl_gr_get_user_gzxa020()
#                  RETURNING 回传参数
# Input parameter: 
# Return code....: gzxa020
# Date & Author..: 190116 By 02716
# Modify.........:#190117-00039#1 add
################################################################################
PUBLIC FUNCTION cl_gr_get_user_gzxa020()
   DEFINE l_gzxa020   LIKE  gzxa_t.gzxa020

    LET l_gzxa020 = ""
    
   SELECT gzxa020 INTO l_gzxa020
   #PGS-11084 add -S-
   #PGS-11084 add -E-
     FROM gzxa_t
    WHERE gzxaent = g_enterprise and gzxa003 = g_user
      AND gzxastus ='Y' 

    IF cl_null(l_gzxa020) THEN
       LET l_gzxa020 ="N"
    END IF
    
    RETURN l_gzxa020
END FUNCTION

################################################################################
# Descriptions...: 判斷是否列印條碼
# Memo...........:
# Usage..........: cl_gr_chk_code()
#                  RETURNING r_flag
# Input parameter: 
#                : 
# Return code....: r_flag      回傳是否列印條碼flag
#                :
# Date & Author..: 2019/02/22 By 11084
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_chk_code()
   DEFINE r_flag           LIKE  type_t.chr1
   DEFINE l_gzgs004        LIKE  gzgs_t.gzgs004
   
   SELECT gzgs004 INTO l_gzgs004 FROM gzgs_t WHERE gzgs001 = g_rep_code AND gzgsstus = 'Y'
        
   IF cl_null(l_gzgs004) THEN
      LET r_flag = cl_get_para(g_enterprise,g_site,"A-SYS-0080")
   ELSE
      LET r_flag = l_gzgs004
   END IF
   
   IF r_flag = "N" THEN
      LET g_code_value = NULL
   END IF
   
   RETURN r_flag
      
END FUNCTION

################################################################################
# Descriptions...: 無報表資料產生時，寄送mail通知
# Memo...........:
# Usage..........: cl_gr_nodata_sendmail()
# Input parameter: 
#                : 
# Return code....: 
#                : 
# Date & Author..: 2019/05/14 By 11084
# Modify.........: #190223-00010#1 add
################################################################################
PUBLIC FUNCTION cl_gr_nodata_sendmail()
   DEFINE ls_file       STRING
   DEFINE ls_text       STRING
   DEFINE ls_tmp        STRING
   DEFINE l_time        STRING
   DEFINE ls_pid        STRING
   DEFINE ls_sql        STRING
   DEFINE l_ch          base.Channel
   DEFINE l_cnt         LIKE type_t.num5
   DEFINE l_ac          LIKE type_t.num5
   DEFINE l_for_ac      LIKE type_t.num5
   DEFINE l_mail_info   RECORD 
             recipient  STRING,
             cc         STRING,
             bcc        STRING
          END RECORD   
   DEFINE l_send_mail   DYNAMIC ARRAY OF RECORD
             gzpe004    LIKE gzpe_t.gzpe004,    #收件人員工編號
             gzpe007    LIKE gzpe_t.gzpe007,    #形式
             gzpe012    LIKE gzpe_t.gzpe012,    #聯絡對象類型
             gzpe013    LIKE gzpe_t.gzpe013     #進階設定
             ,gzpe005   LIKE gzpe_t.gzpe005     #信件主旨   #200110-00003 add
          END RECORD
   
   
   #撈取收件人筆數
   LET ls_sql = "SELECT COUNT(1) FROM gzpe_t WHERE gzpe001 = '",g_schedule_id,"' AND gzpe002 = '",g_schedule_seq,"' AND gzpe003 = 'Z'"
   PREPARE sql_pre FROM ls_sql
   EXECUTE sql_pre INTO l_cnt
   
   #撈取排程信件清單
   LET ls_sql = "SELECT gzpe004,gzpe007,gzpe012,gzpe013 ",
                ",gzpe005 ",   #200110-00003 add
                "FROM ds.gzpb_t ",
                "LEFT JOIN ds.gzpe_t ",
                "ON gzpb001 = gzpe001 ",
                "AND gzpb002 = gzpe002 ",
                "WHERE gzpb001 = '",g_schedule_id,"' ",
                "AND gzpb002 = '",g_schedule_seq,"' ",
                "AND gzpe003 = 'Z'"
   PREPARE sql_pre2 FROM ls_sql
   DECLARE sql_cur CURSOR FOR sql_pre2
   LET l_for_ac = 1
   FOREACH sql_cur INTO l_send_mail[l_for_ac].*
      LET l_for_ac = l_for_ac + 1
   END FOREACH

   #如果沒有填寫任何寄信資料的話返回
   IF l_send_mail.getLength() = 0 THEN
      RETURN
   END IF

   LET l_time = CURRENT HOUR TO FRACTION(5)
   LET l_time = cl_replace_str(EXTEND(l_time, HOUR TO FRACTION(5)), ":", "")
   LET l_time = cl_replace_str(l_time, ".", "_")
   LET ls_pid= FGL_GETPID(),l_time
   LET ls_file = FGL_GETENV("TEMPDIR"),os.Path.separator(),"rep_no_data_",ls_pid.trim(),".txt"
   LET ls_text = "無報表資料"
   LET l_ch =  base.Channel.create()
   CALL l_ch.openFile(ls_file,"a")
   CALL l_ch.writeLine(ls_text)
   FOR l_ac = 1 TO l_cnt 
#      LET g_xml.subject = cl_gr_mail_subject()
      #200110-00003#1 add(s)
      IF l_send_mail[l_ac].gzpe005 = "default" THEN
         LET g_xml.subject = cl_gr_mail_subject()
      ELSE
         LET g_xml.subject = l_send_mail[l_ac].gzpe005,"  SCHEDULE ID: ",g_schedule_id
      END IF
      #200110-00003#1 add(e)
      #信件本文
      LET g_xml.body = ls_file
      #寄信人
      IF cl_null(g_xml.sender) THEN
         CALL cl_gr_get_mailsender()
      END IF
      #收件者
      CASE 
         WHEN l_send_mail[l_ac].gzpe007 = "4"  #4:進階設定
             CALL util.JSON.parse(l_send_mail[l_ac].gzpe013,l_mail_info)
             LET g_xml.recipient = l_mail_info.recipient
             LET g_xml.cc = l_mail_info.cc
             LET g_xml.bcc = l_mail_info.bcc
             LET ls_tmp = "e-mail:",g_xml.recipient ," ",g_xml.cc ," ",g_xml.bcc
             DISPLAY "INFO:寄件資訊:",ls_tmp
         WHEN l_send_mail[l_ac].gzpe007 = "2"  #2:副本
             LET g_xml.cc = cl_s_aooi350_get_cont_ppl_email(l_send_mail[l_ac].gzpe012,l_send_mail[l_ac].gzpe004)
             LET ls_tmp = "e-mail:",g_xml.cc
             DISPLAY "INFO:寄件資訊:",ls_tmp
         WHEN l_send_mail[l_ac].gzpe007 = "3"  #3:密件副本
             LET g_xml.bcc = cl_s_aooi350_get_cont_ppl_email(l_send_mail[l_ac].gzpe012,l_send_mail[l_ac].gzpe004)
             LET ls_tmp = "e-mail:",g_xml.bcc
             DISPLAY "INFO:寄件資訊:",ls_tmp
         OTHERWISE                             #1:收件者 
            LET g_xml.recipient = cl_s_aooi350_get_cont_ppl_email(l_send_mail[l_ac].gzpe012,l_send_mail[l_ac].gzpe004)   
            LET ls_tmp = "e-mail:",g_xml.recipient
            DISPLAY "INFO:寄件資訊:",ls_tmp
      END CASE

      #寄發mail
      CALL cl_jmail() RETURNING ls_tmp
      DISPLAY "cl_jmail: ",ls_tmp
      
   END FOR
   
END FUNCTION

################################################################################
# Descriptions...: 娜娜接口呼叫，產生 PDF 查看網址
# Memo...........:
# Usage..........: CALL cl_gr_athena_view ()
#                  RETURNING PDF URL
# Input parameter: 
# Return code....: l_url   PDF URL
# Date & Author..: #250912-00022#1 add By 14910
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_gr_athena_view()
   DEFINE l_url  STRING

   LET l_url = "http://",cl_rpt_get_env_global("GRIP"),"/",FGL_GETENV("ZONE"),"/temp/gr/",FGL_GETENV("ATHENA_FILE")

   RETURN l_url
END FUNCTION

 
{</section>}
 
