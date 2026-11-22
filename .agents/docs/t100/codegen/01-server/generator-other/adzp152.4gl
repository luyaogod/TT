# code_version.: 5.25.02-2012.04.16(0000-0001)

IMPORT os
SCHEMA ds

#+ code_id......: adzp152.4gl
#+ descriptions.: 程式資料建立
#+ code_creator.: hjwang
# Modify........: No:YSC-E30001 14/03/19 By joyce 新增q04樣板(柏榕修改部分追版)
# Modify........: No:YSC-E30002 14/03/21 By joyce 調整程式可處理有左方QBE的部分
# Modify........: No:YSC-E40001 14/04/15 By joyce 調整q03樣板(與q01共用樣板)
# Modify........: No:YSC-E40004 14/04/21 By joyce wss模組應讀取COM路徑下設定
# Modify........: No:YSC-E40002 14/04/23 By joyce 新增PR版次取用
# Modify........: No:YSC-E40007 14/04/25 By joyce 1.單身b_fill及FOREACH段需濾除掉串查欄位
#                                                 2.單身串查欄位須定義為STRING型態
# Modify........: No:YSC-E50003 14/05/16 By joyce 單頭串查功能調整
# Modify........: No:YSC-E60001 14/06/17 By joyce 增加判斷是否為行業別程式
# Modify........: No:YSC-E70001 14/07/09 By joyce sel欄位改為可動態設定
# Modify........: No:YSC-E70002 14/07/21 By joyce 程式非第一版次不可覆蓋add-point
# Modify........: No:YSC-E80001 14/08/04 By joyce 因應解開section後如果又進行規格調整可能發生樣板對應上的錯誤
# Modify........: No:150529-00031 15/06/01 By joyce 單身串查功能依據環境變數決定要以何種方式呈現
# Modify........: No:170306-00025 17/03/06 By joyce 調整產生mask程式段的擺放位置，讓不論是否解開section的程式都能產生mask段落
# Modify........: No:190108-00038 19/01/18 By 00824 畫面簡化功能
#                                 註：因為簡易畫面是透過設定而來，所以有可能標準化面已客製，但簡易畫面仍為標準的狀況 
#                                     所以簡易畫面開啟時，需以dzzc_t中紀錄的實際模組(dzzc004)為主
# Modify........: No:190826-00045 10/10/09 By 02667 修正參照欄位型態與長度
# Modify........: No:200210-00032 20/02/10 By 00824 調整adzp152_define_table_name中欄位參照的方式
# Modify........: No:200804-00006 20/08/05 By 00824 postgre功能回追
# Modify........: No:201109-00009 20/12/18 By 00824 模糊查詢功能

GLOBALS "../../cfg/top_global.inc"
 
GLOBALS
   DEFINE g_properties   om.SaxAttributes
   CONSTANT li_space = 3
   DEFINE g_component_list DYNAMIC ARRAY OF RECORD
            comp_id    STRING,
            action     LIKE type_t.num5,
            standard   LIKE type_t.num5,
            including  LIKE type_t.num5
            END RECORD
   DEFINE gdnode_all     om.DomNode
# YSC-E30002 ---start---
DEFINE g_form_field DYNAMIC ARRAY OF RECORD
         type         LIKE type_t.chr10,
         page         LIKE type_t.num5,
         sr_record    LIKE type_t.chr20,
         tot_count    LIKE type_t.num5,
         var_field    LIKE type_t.chr50,
         form_field   LIKE type_t.chr50
         END RECORD
# YSC-E30002 --- end ---
DEFINE g_update       STRING   #YSC-E50004 modify
DEFINE g_simple_type  LIKE type_t.chr1   #190108-00038
END GLOBALS

DEFINE g_dzbb RECORD 
         prog_name    STRING,
         point_name   STRING,
         point_ver    STRING,
         addpoint     STRING
         END RECORD
        
DEFINE g_master_chk   BOOLEAN
DEFINE g_detail_chk   BOOLEAN

DEFINE g_value        STRING

CONSTANT cs_prog_ver = 1.0

MAIN
   DEFINE lddoc_all      om.DomDocument
   DEFINE ls_filename    STRING
   DEFINE ls_prog        STRING
   DEFINE ls_path        STRING
   DEFINE ls_cmd         STRING
   DEFINE ls_module      STRING   #YSC-E40004
   DEFINE ls_ver         LIKE type_t.num10   #YSC-E70002
   DEFINE lr_dzbc        RECORD LIKE dzbc_t.*   #YSC-E80001
   DEFINE li_cnt         LIKE type_t.num10      #YSC-E80001
   DEFINE ls_wmode       STRING #ka0132
   
   #定義在其他link的程式則無效
   WHENEVER ERROR CALL cl_err_msg_log

   CALL cl_db_connect("ds",FALSE)

   #是否覆蓋原有add-point
   LET g_update = ARG_VAL(3)
   IF cl_null(g_update) THEN
      LET g_update = 'n'
   END IF

   #YSC-E80001 ---start---
   #顯示外部參數訊息
   DISPLAY "開始執行程式產生器, 以下為產生參數:"
   DISPLAY " 1. 模組名稱:",ARG_VAL(1)
   DISPLAY " 2. 程式名稱:",ARG_VAL(2)
   DISPLAY " 3. 程式add-point是否覆蓋:",g_update
   DISPLAY " 4. 程式版次:",ARG_VAL(4)
   DISPLAY " 5. 程式客製標示:",ARG_VAL(5)
   DISPLAY " 6. 產出模式:",ARG_VAL(6) #ka0132

   #先確認Section是否解開
   LET li_cnt = 0
   LET lr_dzbc.dzbc001 = ARG_VAL(2)
   LET lr_dzbc.dzbc002 = ARG_VAL(4)
   LET lr_dzbc.dzbc007 = ARG_VAL(5)
   SELECT COUNT(*) INTO li_cnt FROM dzbc_t
    WHERE dzbc001 = lr_dzbc.dzbc001  #程式名稱
      AND dzbc002 = lr_dzbc.dzbc002  #PR版次
      AND dzbc005 <> 's'             #使用標示
      AND dzbc007 = lr_dzbc.dzbc007
   
   #ka0132 - Start -
   #檢查6 - section處理模式
   LET ls_wmode = ARG_VAL(6)
   IF cl_null(ls_wmode) THEN
      LET ls_wmode = '1'
   END IF
   IF NOT (ls_wmode = '1'  OR 
           ls_wmode = '2'  OR
           ls_wmode = '3') THEN
      DISPLAY "ERROR(參數6):傳入的section處理模式不符規定, 須為1或2或3!"
      EXIT PROGRAM
   END IF 
   IF ls_wmode = '2' THEN
      DISPLAY '進行預覽模式4gl2產出!'
   END IF
   #ka0132 - End -
	  
   #先判斷是否解開section, 進行解析與寫入
   IF li_cnt > 0 AND ls_wmode = '1' THEN #ka0132
      DISPLAY '該程式(',lr_dzbc.dzbc001,')已解開Section調整, 不再產生標準Section!'
   ELSE
   #YSC-E80001 --- end ---

      #ka0132 - Start -
      #若未解開Section則強制將ls_wmode改為1 
      IF li_cnt = 0 AND ls_wmode <> '1' THEN
         LET ls_wmode = '1'
         DISPLAY "WARNING:該程式未解開Section, section處理模式改正為1!"
      END IF
      #ka0132 - End -
   
      #判斷程式類型
      LET ls_module = ARG_VAL(1)   #YSC-E40004
      IF ls_module = 'lib' OR ls_module = 'sub' OR ls_module = 'wss'   #YSC-E40004
         OR ls_module = 'clib' OR ls_module = 'csub' OR ls_module = 'cwss' THEN   #YSC-E80001 add
 
         LET ls_path = "COM"
      ELSE
         LET ls_path = "ERP"
      END IF
 
      #tabx檔更正為tab檔
      LET ls_filename = os.Path.join(os.Path.join(FGL_GETENV(ls_path),ARG_VAL(1)),"dzx")
      LET ls_filename = os.Path.join(ls_filename,"tab")
      LET ls_filename = os.Path.join(ls_filename,ARG_VAL(2) CLIPPED||".tab")
 
      #YSC-E70002 ---start---
      LET ls_ver = ARG_VAL(4)
      #僅有第一版次能夠覆蓋add-point
      IF g_update = 'y' AND ls_ver > 1 THEN
         LET g_update = 'n'
         DISPLAY "WARNING:該程式非第一版次, 無法覆蓋add-point!"
      END IF
      #YSC-E70002 --- end ---
 
      #定義轉換用SaxAttributes
      LET g_properties = om.SaxAttributes.create()
      CALL g_properties.addAttribute("app_id",ARG_VAL(2)) #app_id 程式代號
      CALL g_properties.addAttribute("module",ARG_VAL(1)) #module 
 
      LET lddoc_all = om.DomDocument.createFromXmlFile(ls_filename)  #om.DomDocument
 
      #若讀取不到資料代表tabx不存在
      TRY
         LET gdnode_all = lddoc_all.getDocumentElement()                #om.DomNode
         CALL g_component_list.clear()
      CATCH
         DISPLAY "ERROR(1):讀取",ls_filename,"發生錯誤, 程式中止!"
         EXIT PROGRAM
      END TRY
 
      LET gdnode_all = lddoc_all.getDocumentElement()                #om.DomNode
      CALL g_component_list.clear()
 
      #產生流程=> 讀取樣板檔案 / 同時間寫出(一進一出) / 產生器符號中斷,擺放切片或產生code
      #獲取程式基本資料 (含填寫待轉換的對照表)
      DISPLAY "程式產生器(adzp152):開始讀取tab並產生基礎資料!"
      CALL adzp152_read_basic_data()
 
      #開始進行產生/轉換動作
      DISPLAY "程式產生器(adzp152):開始讀取template並取代相關變數產生tgl!"
      CALL adzp152_read_and_replace()
      
      
      #將add-point版本一次更新
      CALL cl_add_point_update(g_properties.getValue("app_id"),g_properties.getValue("general_adp_ver"))   #YSC-E60004 modify


      #呼叫adzp175重新產生
      DISPLAY "--------------------------------------------------"
      DISPLAY "開始section解析, 執行adzp175"
      LET ls_cmd = ' adzp175 ',ARG_VAL(1),' ',ARG_VAL(2)
      CALL cl_cmdrun_wait(ls_cmd)
 
      #呼叫adzp176寫入section
      DISPLAY "--------------------------------------------------"
      DISPLAY "開始section解析, 執行adzp176"
      LET ls_cmd = ' adzp176 ',ARG_VAL(1),' ',ARG_VAL(2),' ',ARG_VAL(4),' ',ARG_VAL(5),' ',ls_wmode #ka0132   #YSC-E40002 add ARG_VAL(4)   #YSC-E60001 add ARG_VAL(5)
      CALL cl_cmdrun_wait(ls_cmd)
   END IF

   #呼叫adzp177產出tgl
   DISPLAY "--------------------------------------------------"
   DISPLAY "開始產生tgl檔案, 執行adzp177"
   LET ls_cmd = ' adzp177 ',ARG_VAL(1),' ',ARG_VAL(2),' ',ARG_VAL(4),' ',ARG_VAL(5),' ',ls_wmode #ka0132   #YSC-E40002 add ARG_VAL(4)   #YSC-E60001 add ARG_VAL(5)
   CALL cl_cmdrun_wait(ls_cmd)

   #170306-00025 ---start---
   #呼叫adzp155，產生遮罩功能用的4gl
   DISPLAY "--------------------------------------------------"
   DISPLAY "開始產生mask檔案, adzp155開始運作時間:",cl_get_current()
   LET ls_cmd = ' adzp155 ',ARG_VAL(1),' ',ARG_VAL(2), ' n '
   CALL cl_cmdrun_wait(ls_cmd)
   #170306-00025 --- end ---

   DISPLAY "--------------------------------------------------"

END MAIN


#+ 取出程式基本資料 <assembly>
PRIVATE FUNCTION adzp152_read_basic_data()

   DEFINE ldnode_action om.DomNode
   DEFINE lnl_selted    om.NodeList
   DEFINE ls_err        STRING
   DEFINE ls_tmp        STRING
   DEFINE ls_type_list  STRING
   DEFINE ls_gzza001    LIKE gzza_t.gzza001
   DEFINE li_tmp        LIKE type_t.num5     #YSC-E60001
   DEFINE ls_app        LIKE gzza_t.gzza001  #YSC-E60001
   DEFINE ls_app_s      STRING               #YSC-E60001
   
   #取出om.NodeList 有關tag list
   LET lnl_selted = gdnode_all.selectByTagName("assembly")
   LET ldnode_action = lnl_selted.item(1) 

   #模組
   LET ls_tmp = ldnode_action.getAttribute("module") 
   CALL g_properties.addAttribute("general_module",ls_tmp)
   
   #tiptop版本
   LET ls_tmp = ldnode_action.getAttribute("tpver") 
   CALL g_properties.addAttribute("general_prog_ver",ls_tmp)

   # YSC-E40002 ---start---
   #程式版本(PR)
   CALL g_properties.addAttribute("general_pr_ver",ARG_VAL(4))
   # YSC-E40002 --- end ---

   #程式版本(SD)
   LET ls_tmp = ldnode_action.getAttribute("sdver")
   #YSC-E40002 ---start---
   IF cl_null(ls_tmp) THEN
      LET ls_tmp = "1"
   END IF
   #YSC-E40002 --- end ---
   CALL g_properties.addAttribute("general_sd_ver",ls_tmp)

   #YSC-E60004 ---start---
   #add_point版次
   CALL g_properties.addAttribute("general_adp_ver",ARG_VAL(4))
   #YSC-E60004 --- end ---

   #程式描述
   LET ls_tmp = ldnode_action.getAttribute("description") 
   CALL g_properties.addAttribute("general_title_desc",ls_tmp)
   
   #程式創造者
   LET ls_tmp = ldnode_action.getAttribute("crtid") 
   CALL g_properties.addAttribute("general_title_creator",ls_tmp)
   
   #程式創造時間
   LET ls_tmp = ldnode_action.getAttribute("crtdt") 
   CALL g_properties.addAttribute("general_title_create_time",ls_tmp)
   
   #程式修改者
   LET ls_tmp = ldnode_action.getAttribute("modid") 
   CALL g_properties.addAttribute("general_title_modifier",ls_tmp)
   
   #程式修改時間
   LET ls_tmp = ldnode_action.getAttribute("moddt") 
   CALL g_properties.addAttribute("general_title_modify_time",ls_tmp)
   
   #模組
   LET ls_tmp = ldnode_action.getAttribute("jobmode") 
   CALL g_properties.addAttribute("general_jobmode",ls_tmp)

   #行業別
   LET ls_tmp = ldnode_action.getAttribute("industry") 
   CALL g_properties.addAttribute("industry_id",ls_tmp)

   #form代碼
   LET ls_tmp = ldnode_action.getAttribute("form") 
   CALL g_properties.addAttribute("form_id",ls_tmp)

   #預設參數數量
   LET ls_tmp = ldnode_action.getAttribute("fix_arg")
   CALL g_properties.addAttribute("fix_arg",ls_tmp)
   IF cl_null(ls_tmp) THEN
      LET ls_tmp = '0'
   END IF
   CALL adzp152_arg_idx(ls_tmp)

   #樣板代碼
   LET ls_type_list = "q01,q02,q03,q04,p00,p01,n00"    #YSC-E30001 add q04
   LET ls_tmp = ldnode_action.getAttribute("type")
   IF ls_type_list.getIndexOf(ls_tmp,1) = 0 THEN
      DISPLAY "ERROR(2):目前尚未支持此種樣板檔(",ls_tmp,"), 請重新定義!"
      EXIT PROGRAM
   END IF

   #名義樣板名稱
   CALL g_properties.addAttribute("type_id_t",ls_tmp)
   
   #YSC-E40001 ---start---
   # 目前q03與q01共用q01樣板
   IF ls_tmp = "q03" THEN
      LET ls_tmp = "q01"
   END IF
   #YSC-E40001 --- end ---

   #實際使用的樣板名稱
   CALL g_properties.addAttribute("type_id",ls_tmp)
   
   DISPLAY "目前使用的程式樣板為 : code_",ls_tmp,"樣板"

   #樣板會使用到單身多頁籤,用CASE可避免忘記新增樣板須調整的程式段
   CASE
      WHEN ls_tmp.equals("q01") OR ls_tmp.equals("q02") OR ls_tmp.equals("q03") OR ls_tmp.equals("q04")   #YSC-E30001 add q04  #YSC-E30002 add q01
         #頁簽數 
         LET ls_tmp = ldnode_action.getAttribute("page") 
         CALL g_properties.addAttribute("page",ls_tmp)

   END CASE

   #確認MD5是否正確(錯誤則中斷)
#  CALL sadzi100_md5_chk(g_properties.getValue("type_id"))
   CALL sadzi100_template_chk(g_properties.getValue("type_id"),cs_prog_ver)

   #先取出master_tbl_name跟detail_tbl_name的值,供轉換時應用
   CALL adzp152_get_table_name(ldnode_action)

   #150529-00031 ---start---
   #因為adzp152_get_table_name只取第一個table的資料，
   #但後續仍有其他需求需要取得dataset中各單頭與單身的資訊，所以另新增FUNCTION處理
   CALL adzp152_get_dataset_table_info(ldnode_action)
   #150529-00031 --- end ---

   #YSC-E40007 ---start---
   #先取出有做串查功能的欄位，以便後續確認是否需要濾除或做額外處理
   CALL adzp152_get_hyper_fields(ldnode_action)
   #YSC-E40007 --- end ---


   #YSC-E60001 ---start---
   #判斷是否為行業別程式, 再決定general_prefix
   LET ls_app_s = g_properties.getValue("app_id")
   IF ls_app_s.getIndexOf('_',1) > 0 THEN
      LET ls_app = ls_app_s.subString(1,ls_app_s.getIndexOf('_',1)-1)
      LET li_tmp = 0
      SELECT COUNT(*) INTO li_tmp FROM gzza_t
       WHERE gzza001 = ls_app
      IF li_tmp > 0 THEN
         CALL g_properties.addAttribute("general_prefix",ls_app)
      END IF
   END IF

   IF cl_null(g_properties.getValue("general_prefix")) THEN
      CALL g_properties.addAttribute("general_prefix",g_properties.getValue("app_id"))
   END IF
   #YSC-E60001 --- end ---

   #開始走訪XML
   # 刪除temptable
   DROP TABLE adzp152_form_tmp
   DROP TABLE adzp152_para_tmp

   #建立畫面檔設計資料temp table
   CALL adzp152_create_temp_table()
   # YSC-E30002 --- end ---

   CALL adzp152_xml_search(ldnode_action)

   #填寫待轉換的對照表
   CALL adzp152_read_tab()

   #add by joyce 14/03/04
   CALL adzp152_display_saxattribute()
   #add by joyce 14/03/04
   
END FUNCTION

#add by joyce 14/03/04
FUNCTION adzp152_display_saxattribute()
   DEFINE lc_cha        base.Channel
   DEFINE lc_path       STRING
   DEFINE lc_str        STRING
   DEFINE lc_i          LIKE type_t.num5


   LET lc_path = FGL_GETENV("TEMPDIR"),"/adzp152_display_sax.txt"
   LET lc_cha = base.Channel.create()
   CALL lc_cha.setDelimiter("")
   CALL lc_cha.openFile(lc_path, "w")
   FOR lc_i = 1 TO g_properties.getLength()
      IF lc_i > 1 THEN
         LET lc_str = lc_str,"\n"
      END IF
      LET lc_str = lc_str,g_properties.getName(lc_i)," = ",g_properties.getValueByIndex(lc_i)
   END FOR
   CALL lc_cha.write(lc_str)
   CALL lc_cha.close()

END FUNCTION
#add by joyce 14/03/04


#+ 填寫待轉換的對照表
PRIVATE FUNCTION adzp152_read_tab()
   DEFINE l_str_tab_master     STRING 
   DEFINE l_str_tab_detail     STRING 
   DEFINE li_cnt               LIKE type_t.num5 
   DEFINE li_page              LIKE type_t.num5 
   DEFINE ls_column_chk_body   STRING
   DEFINE ls_column_ctrlp_body STRING
   DEFINE ls_page              STRING
   DEFINE ls_tmp               STRING
   DEFINE li_tmp               LIKE type_t.num5 
   DEFINE ls_err               STRING
   DEFINE l_chklist            om.NodeList
   DEFINE l_chknode            om.DomNode
   DEFINE l_nlist              om.NodeList
   DEFINE l_node               om.DomNode
   DEFINE l_chkn               INTEGER
   DEFINE l_buf                base.StringBuffer
   DEFINE l_master_dataset     STRING   #YSC-E30002
   DEFINE l_detail_dataset     STRING   #YSC-E30002
   DEFINE ls_column_filter_ctrlp_body STRING

   #YSC-E60001 ---mark start---
#  #置換部分
#  CALL g_properties.addAttribute("general_prefix",g_properties.getValue("app_id"))
   #YSC-E60001 --- mark end ---
   
   LET ls_tmp = os.Path.join(g_properties.getValue("module"),"42f")
   LET ls_tmp = os.Path.join(ls_tmp,g_properties.getValue("app_id"))
   IF cl_null(ls_tmp) THEN LET ls_err = ls_err," 主畫面屬性組合失敗" END IF
   CALL g_properties.addAttribute("form",ls_tmp)

   #171011-00006 ---start---
   #先行寫入
   LET g_dzbb.prog_name   = g_properties.getValue("app_id")
   LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
   LET g_dzbb.addpoint    = "#Memos"
   LET g_dzbb.point_name  = "global.memo"
   CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))
   #171011-00006 --- end ---

   #ent/site特殊欄位相關處理
   CALL adzp152_append_field()
   
   #取得所有可查詢欄位資料以便後續使用
   LET ls_tmp = ""
   LET l_chklist = gdnode_all.selectByTagName("construct")
   FOR l_chkn = 1 TO l_chklist.getLength()
      LET l_chknode = l_chklist.item(l_chkn)
      LET ls_tmp = ls_tmp,l_chknode.getAttribute("field"),","
   END FOR
   # add by joyce 140312
   IF NOT cl_null(ls_tmp) THEN
      LET ls_tmp = ls_tmp.subString(1, ls_tmp.getLength()-1)
   END IF
   # add by joyce 140312
   CALL g_properties.addAttribute("general_construct_field",ls_tmp)
   LET ls_tmp = ""

   #組refresh用SQL
   LET ls_tmp = adzp152_makesql('refresh')
   IF cl_null(ls_tmp) THEN LET ls_err = ls_err," 屬性refresh sql有問題" END IF
   CALL g_properties.addAttribute("master_refresh",ls_tmp)

   IF NOT cl_null(ls_err) THEN 
      DISPLAY "ERROR(3):",ls_err LET ls_err="" 
   END IF

   #特定樣板的特定段落執行
   IF adzp152_type_decide("m") THEN

      #特殊操作欄位代換
      LET l_str_tab_master = g_properties.getValue("master_tbl_prefix") 
      
      #單頭欄位控制
      CALL g_properties.addAttribute("location","head" )
      
      #stus
      CALL g_properties.addAttribute("master_field_stus", l_str_tab_master||"stus" )
      CALL g_properties.addAttribute("master_var_stus", g_properties.getValue("master_var_title")||"."||l_str_tab_master||"stus")       #${var_stus}
      
      #欄位單頭檢查(含after,before,controlp)
      CALL g_properties.addAttribute("master_fields_check",  adzp152_create_chk_field("0",3,"i") )
      CALL g_properties.addAttribute("master_fields_ctrlp_i",adzp152_create_chk_field("0",3,"ic") )
      CALL g_properties.addAttribute("master_fields_ctrlp",  adzp152_create_chk_field("0",3,"c") )
      CALL g_properties.addAttribute("master_fields_filter_ctrlp",  adzp152_create_chk_field("0",3,"fc") )

      #create inner join for multi master table
      CALL adzp152_master_inner_join()
      
      #reference填充到add point
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_name  = "show.head.reference"
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = g_properties.getValue("master_vars_reference")
      LET l_master_dataset= g_properties.getValue("master_dataset_"||g_properties.getValue("type_id")||"_exist")
   #  CALL adzp152_insert_addpoint(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint)       #YSC-E50004 mark
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))  #YSC-E50004 modify
  
   END IF

   #單身
   IF adzp152_type_decide("d") THEN

      #設定upd_body_fields及cs_body_fields 組單身欄位頁簽
      CALL adzp152_set_upd_body_fields()

      # add by joyce 14/02/25
      # 設定串查欄位相關資訊
      CALL adzp152_detail_write_hyperlink()
      # add by joyce 14/02/25

      #凍結欄位處理
      CALL g_properties.addAttribute("location","body2")
      LET ls_tmp = adzp152_create_chk_field(0,3,"c")
      CALL g_properties.addAttribute("detail_frozen_fields_ctrlp",ls_tmp)   
      
      #單身欄位控制
      CALL g_properties.addAttribute("location","body" )

      LET l_str_tab_detail = g_properties.getValue("detail_tbl_prefix") 
      
      #改以跑迴圈方式處理單身多頁簽方式
      LET li_page = g_properties.getValue("page")
      
      FOR li_cnt = 1 TO li_page
         LET ls_page = li_cnt USING "<<<"
         CALL g_properties.addAttribute("general_current_page",ls_page)   
         LET ls_column_chk_body = "detail_fields_check"
         LET ls_column_ctrlp_body = "detail_fields_ctrlp"
         LET ls_column_filter_ctrlp_body = "detail_fields_filter_ctrlp"
         LET ls_tmp = adzp152_create_chk_field(li_cnt,3,"i")
         CALL g_properties.addAttribute(ls_column_chk_body||ls_page,ls_tmp)   
         LET ls_tmp = adzp152_create_chk_field(li_cnt,3,"ic")
         CALL g_properties.addAttribute(ls_column_ctrlp_body||ls_page||"_i",ls_tmp)    
         LET ls_tmp = adzp152_create_chk_field(li_cnt,3,"c")
         CALL g_properties.addAttribute(ls_column_ctrlp_body||ls_page,ls_tmp)   
         LET ls_tmp = adzp152_create_chk_field(li_cnt,3,"fc")
         CALL g_properties.addAttribute(ls_column_filter_ctrlp_body||ls_page,ls_tmp)   

      #  #YSC-E30002 ---start---
      #  # 因為非q02的樣板不會有dataset的資訊，所以有些資訊必須透過別的方式提供
      #  IF g_properties.getValue("type_id") <> "q02" THEN
      #     LET ls_tmp = adzp152_create_name(li_cnt, "detail_tbl_pages", "<<<")
      #     CALL g_properties.addAttribute(ls_tmp,li_cnt)

      #     CALL adzp152_create_query_entrance(li_cnt)
      #  END IF
      #  #YSC-E30002 --- end ---
      END FOR 
      
      #多table處理
      # YSC-E30002 ---modify start---
      # 目前只有q02樣板一定會設dataset資訊，其餘樣板可能有可能沒有
      LET l_master_dataset= g_properties.getValue("master_dataset_"||g_properties.getValue("type_id")||"_exist")
      LET l_detail_dataset = g_properties.getValue("detail_dataset_"||g_properties.getValue("type_id")||"_exist")
      CASE g_properties.getValue("type_id")
         WHEN "q04"   #只有q04樣板才有單頭
            IF NOT cl_null(l_master_dataset) AND NOT cl_null(l_detail_dataset) 
               AND (l_master_dataset = 'Y' OR l_detail_dataset = 'Y') THEN
               CALL adzp152_create_detail_entrance()
               #CALL adzp152_create_detail2_entrance()
            END IF

         OTHERWISE
            IF NOT cl_null(l_detail_dataset) AND l_detail_dataset = 'Y' THEN
               CALL adzp152_create_detail_entrance()
               #CALL adzp152_create_detail2_entrance()
            END IF
      END CASE
      # YSC-E30002 --- modify end ---
      
      #reference填充到add point
      LET g_dzbb.point_name  = "detail_show.body.reference"
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = g_properties.getValue("detail_vars_reference")
   #  CALL adzp152_insert_addpoint(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint)        #YSC-E50004 mark
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))   #YSC-E50004 modify
      
      FOR li_tmp = 2 TO g_properties.getValue("page")
         LET ls_tmp = li_tmp USING "<<<"
         LET g_dzbb.point_name  = "detail_show.body",adzp152_page_trans(ls_tmp),".reference"
         LET g_dzbb.prog_name   = g_properties.getValue("app_id")
         LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
         LET g_dzbb.addpoint    = g_properties.getValue("detail_vars_reference"||ls_tmp)
      #  CALL adzp152_insert_addpoint(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint)        #YSC-E50004 mark
         CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))   #YSC-E50004 modify
      END FOR
      
   END IF 
   
   #TREE
   IF adzp152_type_decide("t") THEN
   
      #樹狀欄位控制
      CALL g_properties.addAttribute("location","tree" )
      
      #產出reference段落
      LET l_nlist = gdnode_all.selectByTagName("bs_reference")
      FOR li_cnt = 1 TO l_nlist.getLength()
         LET l_node = l_nlist.item(li_cnt)
         CALL adzp152_create_reference(l_node,2) RETURNING ls_tmp
      END FOR
      
      LET l_buf = base.StringBuffer.create()
      LET ls_tmp = g_properties.getValue("browser_fields_all")
      CALL l_buf.append(ls_tmp)
      CALL l_buf.replace(g_properties.getValue("master_field_lid"), g_properties.getValue("master_field_pid"), 0)
      LET ls_tmp = l_buf.toString()
      CALL g_properties.addAttribute("browser_fields_expand", ls_tmp )
     
   END IF
   
   #Browser
   IF adzp152_type_decide("b") THEN
   
      #browser欄位控制
      CALL g_properties.addAttribute("location","browser" )
      
      #產出reference段落
      LET l_nlist = gdnode_all.selectByTagName("bs_reference")
      FOR li_cnt = 1 TO l_nlist.getLength()
         LET l_node = l_nlist.item(li_cnt)
         CALL adzp152_create_reference(l_node,2) RETURNING ls_tmp
      END FOR
      
   END IF

   # YSC-E30001 ---start---
   # 後續qbe的部分會用add-point讓使用者自行處理，所以這段有可能不需要，但先保留
   # q04, qbe欄位框相關
   CALL g_properties.addAttribute("location","qbe" )
   CALL g_properties.addAttribute("qbe_fields_ctrlp",  adzp152_create_chk_field("0",3,"c") )
   CALL g_properties.addAttribute("qbe_fields_ref",  adzp152_create_chk_field("0",3,"i") )
   CALL adzp152_make_qbe()
   # YSC-E30001 --- end ---

   #特別註解處理
   CALL adzp152_mark()
   
   #多語言檔處理(單頭)
   LET ls_tmp = g_properties.getValue("master_fields_lang")
   IF NOT cl_null(ls_tmp) THEN
      CALL g_properties.addAttribute("master_multi_language", adzp152_master_multilang(ls_tmp))
   END IF

   #多語言檔處理(單身)
   LET ls_tmp = g_properties.getValue("detail_fields_lang")
   IF NOT cl_null(ls_tmp) THEN
      CALL g_properties.addAttribute("detail_multi_language", adzp152_detail_multilang(ls_tmp))
   END IF
   
   #190108-00038 ---start---
   #簡易畫面相關處理
   CALL adzp152_simple_mode()
   #190108-00038 --- end ---

   #產生slice相關段落(特定樣板使用)
   CALL adzp152_create_slice()
   
   #state切換資料組成
   CALL adzp152_create_stus()
   
   #公用欄位處理&組成
   CALL adzp152_create_common()

   #遮罩相關處理
   CALL adzp152_create_all_mask()

END FUNCTION



#+ 程式樣板讀取/程式寫出(一進一出) / 產生器符號中斷,擺放切片或產生code
PRIVATE FUNCTION adzp152_read_and_replace()
   DEFINE lchannel_read           base.Channel
   DEFINE lchannel_write          base.Channel
   DEFINE ls_readline             STRING
   DEFINE ls_text                 STRING
   DEFINE ls_code_filename        STRING
   DEFINE ls_sample_filename      STRING
   DEFINE lchannel_check          base.Channel
   DEFINE ls_mdlpath              STRING

   LET lchannel_read = base.Channel.create()
   LET lchannel_write = base.Channel.create()

   CALL lchannel_read.setDelimiter("")
   CALL lchannel_write.setDelimiter("")

   #先讀取T100TEMPLATEPATH, 若無則採用$ERP/mdl
   LET ls_mdlpath = FGL_GETENV("T100TEMPLATEPATH")
   IF cl_null(ls_mdlpath) THEN
      LET ls_mdlpath = FGL_GETENV("ERP")
      LET ls_mdlpath = os.Path.join(ls_mdlpath,"mdl")
   END IF

   #定義取用樣板檔案
   LET ls_sample_filename = "code_",DOWNSHIFT(g_properties.getValue("type_id")),".template"
#  LET ls_sample_filename = os.Path.join(os.Path.join(FGL_GETENV("ERP"),"mdl"),ls_sample_filename)
   LET ls_sample_filename = os.Path.join(ls_mdlpath,ls_sample_filename)
   DISPLAY "樣板檔位置:",ls_sample_filename

   IF NOT os.Path.exists(ls_sample_filename) THEN
      DISPLAY "ERROR(4):樣板檔案:",ls_sample_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   CALL lchannel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #產出程式路徑
   LET ls_code_filename = os.Path.join(FGL_GETENV(UPSHIFT(g_properties.getValue("module"))),"dzx")
   LET ls_code_filename = os.Path.join(ls_code_filename,"tgl")
   LET ls_code_filename = os.Path.join(ls_code_filename,ARG_VAL(2) CLIPPED||".tgx")
   
   #先行移除tgl
   IF os.Path.delete(ls_code_filename) THEN
      DISPLAY "刪除舊檔案:",ls_code_filename
   END IF
   
   #判斷是否砍除成功
   IF NOT os.Path.exists(ls_code_filename) THEN
      DISPLAY "舊檔案刪除成功:",ls_code_filename
   ELSE
      DISPLAY "Error:舊檔案刪除失敗:",ls_code_filename
      EXIT PROGRAM
   END IF
   
   DISPLAY "產生檔位置:",ls_code_filename
   CALL lchannel_write.openFile( ls_code_filename, "w" )

   #產生程式版本及說明
   CALL lchannel_write.write(adzp152_prog_memo())

   #讀取及轉換
   WHILE TRUE
   
      LET ls_text = ""
      LET ls_readline = lchannel_read.readLine()

      #讀取到版本資料時進行回寫gzzx
      IF ls_readline.getIndexOf('樣板自動產生',1) > 0 THEN
         CALL adzp152_update_gzzx(ls_readline)
      END IF

      IF lchannel_read.isEof() THEN
         EXIT WHILE
      END IF

      #產生code部分
      CASE
         #page段落產生
         WHEN ls_readline.getIndexOf("#pages - Start -",1) > 0
            CALL adzp152_make_pages(lchannel_read,lchannel_write,0) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #page段落產生(單身的單頭)
         WHEN ls_readline.getIndexOf("#pages_m - Start -",1) > 0
            CALL adzp152_make_pages(lchannel_read,lchannel_write,5) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #page段落產生(單身的單身)
         WHEN ls_readline.getIndexOf("#pages_d - Start -",1) > 0
            CALL adzp152_make_pages(lchannel_read,lchannel_write,6) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #單頭key段落產生    
         WHEN ls_readline.getIndexOf("#master_keys - Start -",1) > 0
            CALL adzp152_make_keys(lchannel_read,lchannel_write,'m') RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #單身key段落產生    
         WHEN ls_readline.getIndexOf("#detail_keys - Start -",1) > 0
            CALL adzp152_make_keys(lchannel_read,lchannel_write,'d') RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #多table段落產生    
         WHEN ls_readline.getIndexOf("#tables - Start -",1) > 0
            CALL adzp152_make_tables(lchannel_read,lchannel_write,0) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #多table段落產生(主要table)   
         WHEN ls_readline.getIndexOf("#tables_m - Start -",1) > 0
            CALL adzp152_make_tables(lchannel_read,lchannel_write,1) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #多table段落產生(附帶table)
         WHEN ls_readline.getIndexOf("#tables_d - Start -",1) > 0
            CALL adzp152_make_tables(lchannel_read,lchannel_write,2) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #多table段落產生(附帶table的table)
         WHEN ls_readline.getIndexOf("#tables_d2 - Start -",1) > 0
            CALL adzp152_make_tables(lchannel_read,lchannel_write,3) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""

         #迴圈用
         WHEN ls_readline.getIndexOf("#loop_times - Start -",1)
            CALL adzp152_make_loop(lchannel_read,lchannel_write,"A") RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
      END CASE
      
      #行代換/ 對 ${} 置換
      IF ls_readline.getIndexOf("${",1) AND 
         ( ls_readline.getIndexOf("${",1) < ls_readline.getIndexOf("}",1) ) THEN
         LET ls_text = adzp152_line_replace(ls_readline)
      END IF

      #一般未進行任何處理段落產出
      IF ls_text IS NULL THEN
         LET ls_text = ls_readline
      END IF
      CALL lchannel_write.write(ls_text)

   END WHILE 

   #附加被定義的元件
   CALL lchannel_write.write(ls_text)

   CALL lchannel_read.close()
   CALL lchannel_write.close()
   
   #對產生器寫出的檔案權限在UNIX下均全部打開
   IF os.Path.separator() = "/" THEN
      IF os.Path.chrwx(ls_code_filename,511) THEN
      END IF
   END IF
   
END FUNCTION



#+ 產生程式版本及說明
PRIVATE FUNCTION adzp152_prog_memo()
   DEFINE ls_text      STRING
   DEFINE ls_buildtime STRING
   DEFINE li_buildtime INTEGER
   DEFINE ls_value     STRING
   
   LET li_buildtime = adzp152_prog_buildtime()
   LET ls_buildtime = li_buildtime USING "&&&&&&"
    
   #產生次數 
   CALL g_properties.addAttribute("mdl_buildtime", ls_buildtime)
   
   #程式代號
   CALL g_properties.addAttribute("mdl_filename",ARG_VAL(2))
   
   #程式描述
   LET ls_value = g_properties.getValue("general_title_desc")
   CALL g_properties.addAttribute("mdl_desc",ls_value)
   
   #程式創造者
   LET ls_value = g_properties.getValue("general_title_creator")
   CALL g_properties.addAttribute("mdl_creator",ls_value)
   
   #程式創造時間
   LET ls_value = g_properties.getValue("general_title_create_time")
   CALL g_properties.addAttribute("mdl_create_time",ls_value)
   
   #程式修改者
   LET ls_value = g_properties.getValue("general_title_modifier")
   CALL g_properties.addAttribute("mdl_modifier",ls_value)
   
   #程式修改時間
   LET ls_value = g_properties.getValue("general_title_modify_time")
   CALL g_properties.addAttribute("mdl_modify_time",ls_value)
   
   #樣板名稱
   CALL g_properties.addAttribute("mdl_id",g_properties.getValue("type_id"))
   
   LET ls_text = adzp152_make_slice("a00")
   
   RETURN ls_text
   
END FUNCTION



#+ 取出版本建置次數
PRIVATE FUNCTION adzp152_prog_buildtime()
   DEFINE lc_dzaz001   LIKE dzaz_t.dzaz001
   DEFINE li_dzaz002   LIKE dzaz_t.dzaz002
   DEFINE lc_dzaz003   LIKE dzaz_t.dzaz003
   DEFINE lc_dzaz004   DATETIME YEAR TO SECOND

   LET lc_dzaz001 = ARG_VAL(2)
   LET li_dzaz002 = 0
   
   SELECT MAX(dzaz002) INTO li_dzaz002 FROM dzaz_t WHERE dzaz001 = lc_dzaz001
   
   IF STATUS OR li_dzaz002 IS NULL THEN
      LET li_dzaz002 = 0
   END IF
   
   LET lc_dzaz003 = FGL_GETENV("LOGNAME")
   LET lc_dzaz004 = CURRENT

   INSERT INTO dzaz_t(dzaz001,dzaz002,dzaz003,dzaz004)
      VALUES(lc_dzaz001,li_dzaz002+1,lc_dzaz003,lc_dzaz004)

   RETURN li_dzaz002 USING "&&&&"
   
END FUNCTION



#+ 逐行代換
PUBLIC FUNCTION adzp152_line_replace(ls_read)
   DEFINE ls_read     STRING
   DEFINE ls_text     STRING
   DEFINE ls_tag      STRING
   DEFINE li_pos1     LIKE type_t.num10
   DEFINE li_pos2     LIKE type_t.num10
   DEFINE ls_temp     STRING                #暫存properties資料用

   LET li_pos1 = ls_read.getIndexOf("${",1)
   LET li_pos2 = ls_read.getIndexOf("}", li_pos1)

   IF li_pos1 > 0 AND li_pos2 > 0 AND li_pos1 < li_pos2 THEN
      LET ls_text = ""
      LET ls_tag = ls_read.subString(li_pos1 +2, li_pos2 -1 ) #取出要置換的tag

      #由SaxAttribute內取出值進行代換
      #不在行首
      IF li_pos1 > 1 THEN
         LET ls_text = ls_read.subString(1,li_pos1 -1)
      END IF

      #中間段
      LET ls_temp = g_properties.getValue(ls_tag) CLIPPED
         
      LET ls_text = ls_text,g_properties.getValue(ls_tag) CLIPPED,
                       ls_read.subString(li_pos2+1,ls_read.getLength())

      #遞迴處理同行其他組
      IF ls_text.getIndexOf("${",1) THEN
         LET ls_text = adzp152_line_replace(ls_text)
      END IF

   END IF

   RETURN ls_text

END FUNCTION
    


#+ 走訪XML
PRIVATE FUNCTION adzp152_xml_search(ln_n)
   DEFINE ln_n                om.DomNode
   DEFINE ls_tmp              STRING
   DEFINE l_master_dataset    STRING    #YSC-E30002
   DEFINE l_detail_dataset    STRING    #YSC-E30002
   
   LET g_master_chk = FALSE
   LET g_detail_chk = FALSE
   
   #取得第一個子節點 returns the first child node
   LET ln_n = ln_n.getFirstChild()
   WHILE ln_n IS NOT NULL
      CASE ln_n.getTagName()                                 
         WHEN "section"   #取 structure->section
            CASE 
               WHEN ln_n.getAttribute("id") = "global_var"
                  CALL adzp152_xml_global_var_search(ln_n) 

               #產生action的程式段落
               WHEN ln_n.getAttribute("id") = "menu"
                  CALL adzp152_write_on_action(3,g_properties.getValue("type_id"),ln_n,"menu",'')
    
               WHEN ln_n.getAttribute("id") = "detail_show"
                  CALL adzp152_write_on_action(3,g_properties.getValue("type_id"),ln_n,"detail_show",ln_n.getAttribute("page"))

                  # add by joyce 14/03/06
                  # 取串查功能資訊
                  CALL adzp152_get_hyperlink_info(ln_n,"detail_show",ln_n.getAttribute("page"))
                  # add by joyce 14/03/06

               # YSC-E30002 ---start---
               # 建立右方瀏覽畫面的欄位與畫面欄位的對應關係表
               WHEN ln_n.getAttribute("id") = "form_field"
                  CALL adzp152_form_field_search(ln_n)
               # YSC-E30002 --- end ---

               WHEN ln_n.getAttribute("id") = "master_input"
                  CALL adzp152_write_on_action(3,g_properties.getValue("type_id"),ln_n,"master_input",'')

               WHEN ln_n.getAttribute("id") = "detail_input"
                  CALL adzp152_write_on_action(3,g_properties.getValue("type_id"),ln_n,"detail_input",ln_n.getAttribute("page"))

               WHEN ln_n.getAttribute("id") = "construct"
                  CALL adzp152_xml_construct_search(ln_n)

               #產生串查action的程式段落
               #WHEN ln_n.getAttribute("id") = "chain"
               
            END CASE 
            CALL adzp152_xml_search(ln_n)

         WHEN "form"     #取 form 
            IF ln_n.getAttribute("site") = "Y" THEN
               LET ls_tmp = "DISPLAY g_site TO formonly.site"
               CALL g_properties.addAttribute("general_display_site",ls_tmp)
            END IF
            CALL adzp152_xml_search(ln_n)

         WHEN "dataset"  #取 dataset
            # YSC-E30002 ---modify start---
            # 因為除了q02一定會設定dataset資訊外，其餘樣板可能會設也可能不會
            LET l_master_dataset = g_properties.getValue("master_dataset_"||g_properties.getValue("type_id")||"_exist")
            LET l_detail_dataset = g_properties.getValue("detail_dataset_"||g_properties.getValue("type_id")||"_exist")
            CASE g_properties.getValue("type_id")
               WHEN "q04"   #只有q04樣板才有單頭
                  IF NOT cl_null(l_master_dataset) AND NOT cl_null(l_detail_dataset)
                     AND (l_master_dataset = "Y" OR l_detail_dataset = "Y") THEN
                     CALL adzp152_xml_dataset_search(ln_n)
                  END IF

               OTHERWISE
                  IF NOT cl_null(l_detail_dataset) AND l_detail_dataset = "Y" THEN
                     CALL adzp152_xml_dataset_search(ln_n)
                  END IF
            END CASE
            # YSC-E30002 --- modify end ---

         WHEN "init"    #取 form->init
            CASE
               WHEN ln_n.getAttribute("id") = "field_set"
                  CALL  adzp152_xml_field_set_search(ln_n)                 

               #201109-00009 ---start---
               WHEN ln_n.getAttribute("id") = "fuzzyquery"
                  CALL adzp152_xml_fuzzyquery_search(ln_n)
               #201109-00009 --- end ---
            END CASE
            
         OTHERWISE 
            CALL adzp152_xml_search(ln_n)
            
      END CASE  
      LET ln_n = ln_n.getNext()  #同一層 next sibling 
      
   END WHILE 
   
END FUNCTION



#+ 取 structure->section id="global_var" 段落
PRIVATE FUNCTION adzp152_xml_global_var_search(ln_n)

   DEFINE ln_n             om.DomNode
   DEFINE li_page          LIKE type_t.num5
   DEFINE ls_sr_name       STRING 
   DEFINE ls_err           STRING
   DEFINE ls_tmp           STRING
   DEFINE ls_tmp2          STRING
   DEFINE l_tmp            STRING
   DEFINE ls_name          STRING
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token         STRING
   DEFINE ls_bsvars        STRING
   DEFINE ls_page          STRING
   DEFINE ls_id            STRING
   DEFINE l_master_dataset STRING   #YSC-E30002
   DEFINE l_detail_dataset STRING   #YSC-E30002
   
   LET ln_n = ln_n.getFirstChild()
   WHILE ln_n IS NOT NULL 

      LET ls_tmp = ln_n.getAttribute("value"),',', g_properties.getValue("general_form_field")
      
      CALL g_properties.addAttribute("general_form_field", ls_tmp )
   
      CASE ln_n.getAttribute("id") 
         WHEN "bs_field"    #取global_var中的 bs_field,BROWSER欄位定義 
            LET ls_tmp = ln_n.getAttribute("order")
            IF cl_null(ls_tmp) THEN
               LET ls_tmp = ln_n.getAttribute("pk") 
            END IF
            CALL g_properties.addAttribute("bs_order", ls_tmp )
            
            LET ls_tmp = ln_n.getAttribute("value")
            
            IF cl_null(g_properties.getValue("browser_fields_all_o")) THEN
               CALL g_properties.addAttribute("browser_fields_all_o", ln_n.getAttribute("value") )
            ELSE
               CALL g_properties.addAttribute("browser_fields_all_o", g_properties.getValue("browser_fields_all_o")||","||ln_n.getAttribute("value") )
            END IF
            
            CALL adzp152_create_borwser_filter(ls_tmp)
            
            LET ls_page = ln_n.getAttribute("page")
            LET ls_sr_name = ln_n.getAttribute("record")
            LET ls_name = adzp152_create_name(ls_page, "browser_sr_name", "<<<") 
            CALL g_properties.addAttribute(ls_name, ls_sr_name )

            #因應Q類
            LET ls_name  =adzp152_create_name(ls_page, "browser_all_field", "<<<") 
            LET ls_tmp = ln_n.getAttribute("value")
            CALL g_properties.addAttribute(ls_name, ls_tmp )
            
            #去除特別欄位參照內容
            LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
            LET ls_tmp = g_properties.getValue("browser_fields_all")
            IF NOT cl_null(ls_tmp) THEN
               LET ls_tmp = ls_tmp, ","
            END IF
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               IF ls_token.getIndexOf('(',1) > 0 OR
                  ls_token.getIndexOf(g_properties.getValue("master_tbl_prefix"),1) = 0 THEN 
                  LET ls_token = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
                  IF cl_getField(g_properties.getValue("master_tbl_name"),ls_token) THEN
                     LET ls_tmp = ls_tmp, ls_token, ","
                  ELSE
                     LET ls_tmp = ls_tmp, "'',"
                  END IF
                  LET ls_tmp2 = ls_tmp2, ls_token, ","
               ELSE
                  LET ls_tmp = ls_tmp, ls_token, ','
                  LET ls_tmp2 = ls_tmp2, ls_token, ','
               END IF
            END WHILE
            LET ls_tmp = ls_tmp.substring(1,ls_tmp.getLength()-1)
            LET ls_tmp2 = ls_tmp2.substring(1,ls_tmp2.getLength()-1)
            CALL g_properties.addAttribute("browser_fields_insert_tmp", ls_tmp2 )
            CALL g_properties.addAttribute("browser_fields_all", ls_tmp )
    
            LET ls_tmp = ln_n.getAttribute("value")
            LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
            LET ls_bsvars = ""
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               IF ls_token.getIndexOf('(',1) > 0 THEN
                  LET ls_token = ls_token.subString(1, ls_token.getIndexOf('(',1)-1)
               END IF      
               LET ls_bsvars = ls_bsvars,"g_browser[g_cnt].b_",ls_token
               IF lst_token.hasMoreTokens() THEN
                  LET ls_bsvars = ls_bsvars,","
               END IF
            END WHILE               
            CALL g_properties.addAttribute("browser_vars_all", ls_bsvars )
        

            #檢查如果樣板含 Browser 需求，則務必要此項"type_id"
            LET ls_err = g_properties.getValue("type_id")
            
            #若樣板不含browser則不進行以下處理
            IF ls_err.equals("q01") OR ls_err.equals("q03") OR ls_err.equals("q04") THEN   #YSC-E30001 add q04
               #定義 Browser的變數
               IF NOT adzp152_browser_fields_define(g_properties.getValue("browser_fields_all_o"),ln_n.getAttribute("append"),ln_n.getAttribute("pk"),ln_n.getAttribute("ref_table"),ln_n.getAttribute("ref_wc"),ln_n.getAttribute("ref_type")) THEN
                  #DISPLAY "ERROR(5):設定g_browse變數發生問題"
               END IF
            END IF

         #取單頭設定資料    
         WHEN "head"
            #單頭 RECORD 設定,調整命名來源
            LET ls_tmp = g_properties.getValue("master_tbl_prefix")
            LET l_tmp = ls_tmp
            LET ls_tmp = "g_",ls_tmp,"_m"

            CALL g_properties.addAttribute("master_var_title",ls_tmp)

            #YSC-E30001 ---start---
            #(q04)特別處理, 單頭變數名稱
            IF g_properties.getValue("type_id") = 'q04' THEN
               #YSC-E30002 ---modify start---
               LET l_master_dataset = g_properties.getValue("master_dataset_"||g_properties.getValue("type_id")||"_exist")
               IF NOT cl_null(l_master_dataset) AND l_master_dataset = "N" THEN
                  CALL g_properties.addAttribute("master_var_title","g_master")
               END IF
               #YSC-E30002 ---modify start---
            END IF
            #YSC-E30001 --- end ---

            CALL g_properties.addAttribute("title_master",l_tmp)
 
            #master_fields_define及var_h_fill的設定
            CALL adzp152_master_fields_define(li_page,ln_n.getAttribute("value"))
            
            #master_all_field
            CALL g_properties.addAttribute("master_all_field",ln_n.getAttribute("value")) 
            
         #取單身頁面設定資料
         WHEN "body"
            LET li_page = ln_n.getAttribute("page")
            
            LET ls_id = ln_n.getAttribute("page_id")
            IF cl_null(ls_id) THEN
               LET ls_id = ln_n.getAttribute("page")
            END IF
            LET ls_name = adzp152_create_name(li_page, "detail_page_id", "<<<") 
            CALL g_properties.addAttribute(ls_name,ls_id)

            #單身DYNAMIC ARRAY設定
            LET ls_tmp = ln_n.getAttribute("record")
            LET ls_name = adzp152_create_name(li_page, "general_page_input", "<<<") 
            IF ls_tmp.getIndexOf("_info",1) > 0 OR
               ln_n.getAttribute("input") = "N" THEN
               CALL g_properties.addAttribute(ls_name,"N")
            ELSE
               CALL g_properties.addAttribute(ls_name,"Y")
            END IF
            
            #紀錄各個page所含之欄位
            LET l_tmp = li_page USING "<<<"
            CALL g_properties.addAttribute("detail_field_loc"||l_tmp,ln_n.getAttribute("value"))
            
            IF li_page = 1 THEN 
            
               CALL g_properties.addAttribute("sr_name",ls_tmp)  #${sr_name1}

               #調整命名來源
               LET ls_tmp = g_properties.getValue("detail_tbl_prefix")
               LET l_tmp = ls_tmp
               LET ls_tmp = "g_",ls_tmp,"_d"

               CALL g_properties.addAttribute("detail_var_title",ls_tmp)
               CALL g_properties.addAttribute("title_detail",l_tmp)
            ELSE
               LET ls_sr_name = "sr_name",li_page USING "<<<"
               CALL g_properties.addAttribute(ls_sr_name,ls_tmp)  #${sr_name2}

               LET ls_sr_name = "detail_var_title",li_page USING "<<<"
               LET ls_tmp = g_properties.getValue("detail_tbl_prefix")
               LET ls_tmp = "g_",ls_tmp,adzp152_page_trans(li_page), "_d"
               CALL g_properties.addAttribute(ls_sr_name,ls_tmp)
            END IF                                             

            # YSC-E30001 ---start---
            #(q04)特別處理, 單身變數名稱(加上page編號)
            IF g_properties.getValue("type_id") = 'q01' OR g_properties.getValue("type_id") = 'q04' THEN
               #YSC-E30002 ---modify start---
               LET l_detail_dataset = g_properties.getValue("detail_dataset_"||g_properties.getValue("type_id")||"_exist")
               IF NOT cl_null(l_detail_dataset) AND l_detail_dataset = "N" THEN
                  CALL g_properties.addAttribute(adzp152_create_name(li_page, "detail_var_title", "<<<"),
                                                 adzp152_create_name(li_page, "g_detail",         "<<<"))
               END IF
               #YSC-E30002 --- modify end ---
            END IF
            # YSC-E30001 --- end ---

            #detail_fields_define及detail_vars_all的設定
            CALL adzp152_detail_fields_define(li_page,adzp152_field_filter(ln_n.getAttribute("value"),"field"))

            #detail_all_field
            LET ls_tmp = g_properties.getValue("detail_all_field")
            LET ls_tmp = ls_tmp, ln_n.getAttribute("value")
            CALL g_properties.addAttribute("detail_all_field",ls_tmp)

         # YSC-E30002 ---start---
         # 左側qbe相關資訊(非q02樣板)
         WHEN "qbe"
            LET ls_tmp = ln_n.getAttribute("value")
            CALL g_properties.addAttribute("qbe_field", ls_tmp )

            LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               IF NOT ls_token.getIndexOf("(",1) THEN  #要避開第一個欄位是FORMONLY的欄位
                  CALL g_properties.addAttribute("qbe_first_field", ls_token )
                  EXIT WHILE
               END IF
            END WHILE
         # YSC-E30002 --- end ---

         OTHERWISE 
            IF ln_n.getTagName() <> "bs_reference" THEN 
               DISPLAY "WARNING: global_var區段存在不允許的屬性:",ln_n.getAttribute("id") 
            END IF
      END CASE 
      LET ln_n = ln_n.getNext()  #同一層 next sibling   
      
   END WHILE

END FUNCTION 


#+ CONSTRUCT段落讀取資料
PRIVATE FUNCTION adzp152_xml_construct_search(n)

   DEFINE n          om.DomNode
   DEFINE l_page     LIKE type_t.num5
   DEFINE l_str_page STRING 
   DEFINE ls_tmp     STRING 
   
   LET n = n.getFirstChild()
   
   WHILE n IS NOT NULL
      
      CASE 
         WHEN n.getAttribute("id")  = "head" #${master_fields_qbe}
            LET ls_tmp = g_properties.getValue("type_id")
            CALL g_properties.addAttribute("master_fields_qbe",adzp152_fields_qbe_filter(g_properties.getValue("master_all_field")))
            CALL adzp152_browser_filter_define(n.getAttribute("page"),n.getAttribute("field"))
            
         WHEN n.getAttribute("id")  = "body"
            LET l_page = n.getAttribute("page")

            LET l_str_page = l_page
            IF l_page = 1 THEN
               LET ls_tmp = g_properties.getValue("sr_name")
            ELSE
               LET ls_tmp = g_properties.getValue("sr_name"||l_str_page)
            END IF
            CALL g_properties.addAttribute("detail_fields_qbe"||l_str_page,n.getAttribute("field")) 
            CALL g_properties.addAttribute("detail_srfield_all"||l_str_page,adzp152_change_cs_sr_detail(ls_tmp,n.getAttribute("field")))

         # YSC-E30001 ---start---
         WHEN n.getAttribute("id")  = "qbe"
            CALL g_properties.addAttribute("qbe_all_fields",n.getAttribute("field"))
         # YSC-E30001 --- end ---
      END CASE 
      LET n = n.getNext()  #同一層 next sibling 
   END WHILE 
END FUNCTION 

#201109-00009 ---start---
#列出所有可做模糊查詢的欄位list
PRIVATE FUNCTION adzp152_xml_fuzzyquery_search(pn_node)
   DEFINE pn_node         om.DomNode
   DEFINE li_page         LIKE type_t.num5
   DEFINE ls_str_page     STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING


   LET pn_node = pn_node.getFirstChild()

   WHILE pn_node IS NOT NULL

      CASE
         WHEN g_properties.getValue("type_id") = "q01" OR g_properties.getValue("type_id") = "q04"
              OR g_properties.getValue("type_id") = "q01"
              CALL g_properties.addAttribute("master_fuzzyquery_fields",pn_node.getAttribute("field"))

         WHEN g_properties.getValue("type_id") = "q02"
            IF pn_node.getAttribute("id") = "body" THEN
               LET li_page = pn_node.getAttribute("page")
               LET ls_str_page = li_page
               IF li_page = 1 THEN
                  CALL g_properties.addAttribute("detail_fuzzyquery_fields",pn_node.getAttribute("field"))
               ELSE
                  LET ls_value = pn_node.getAttribute("field")
                  LET ls_name = adzp152_create_name(li_page, "detail_fuzzyquery_fields", "<<<")
                  CALL g_properties.addAttribute(ls_name,ls_value)
               END IF
            END IF
      END CASE

   #  CASE
   #     WHEN pn_node.getAttribute("id") = "head"
   #        CALL g_properties.addAttribute("master_fuzzyquery_fields",pn_node.getAttribute("field"))

   #     WHEN pn_node.getAttribute("id") = "body"
   #        LET li_page = pn_node.getAttribute("page")
   #        LET ls_str_page = li_page
   #        IF li_page = 1 THEN
   #           CALL g_properties.addAttribute("detail_fuzzyquery_fields",pn_node.getAttribute("field"))
   #        ELSE
   #           LET ls_value = pn_node.getAttribute("field")
   #           LET ls_name = adzp152_create_name(li_page, "detail_fuzzyquery_fields", "<<<")
   #           CALL g_properties.addAttribute(ls_name,ls_value)
   #        END IF
   #  END CASE
      LET pn_node = pn_node.getNext()  #同一層 next sibling

   END WHILE

END FUNCTION
#201109-00009 --- end ---

PRIVATE FUNCTION adzp152_xml_field_set_search(n)
   DEFINE n om.DomNode
   
   LET n = n.getFirstChild()
   
   WHILE n IS NOT NULL
      
      CASE 
         WHEN n.getAttribute("id") = "head" AND n.getAttribute("type") = "entry"  #${entry}
            CALL g_properties.addAttribute("master_entry_fields",n.getAttribute("field"))
         WHEN n.getAttribute("id")  = "head" AND n.getAttribute("type")  = "no_entry" #${no_entry}
            CALL g_properties.addAttribute("master_no_entry_fields",n.getAttribute("field"))   
      END CASE 
      LET n = n.getNext()  #同一層 next sibling 
      
   END WHILE 
   
END FUNCTION 


PRIVATE FUNCTION adzp152_xml_dataset_search(n)
   DEFINE n                   om.DomNode
   DEFINE n2                  om.DomNode
   DEFINE l_tag_attribute     STRING
   DEFINE ls_fix              STRING
   DEFINE li_idx              INTEGER
   DEFINE ls_idx              STRING
   DEFINE ls_tmp              STRING
   DEFINE ls_tmp2             STRING
   DEFINE ls_name             STRING
   DEFINE ls_value            STRING
   DEFINE ls_field_allkeys    STRING
   DEFINE ls_var_allkeys      STRING
   DEFINE lst_token           base.StringTokenizer
   DEFINE ls_token            STRING
   
   #先行取得第三階單身資料
   LET n2 = n.getFirstChild()
   
   LET n = n.getFirstChild()  #取得第一個子節點 returns the first child node
   WHILE n IS NOT NULL
      CASE n.getTagName()
         WHEN "head"
            CALL g_properties.addAttribute("master_tbl_cnt",1)
            
            IF g_master_chk THEN
               CALL adzp152_master_multitable_info(n)
               LET n = n.getNext()
               CONTINUE WHILE
            END IF
            
            CALL g_properties.addAttribute("master_tbl_name_by_tbl",n.getAttribute("id") )
            
            LET g_master_chk = TRUE
         
            LET ls_field_allkeys = ""
            LET ls_var_allkeys = ""
            
            #TREE lid,pid,type定義判斷
            LET ls_tmp = n.getAttribute("lid")
            LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
            IF ls_tmp2.getIndexOf(ls_tmp,1) THEN
               CALL g_properties.addAttribute("master_lid_define_mark","#")
            END IF
            LET ls_tmp = n.getAttribute("pid")
            LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
            IF ls_tmp2.getIndexOf(ls_tmp,1) THEN
               CALL g_properties.addAttribute("master_pid_define_mark","#")
            END IF
            LET ls_tmp = n.getAttribute("type")
            LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
            IF ls_tmp2.getIndexOf(ls_tmp,1) THEN
               CALL g_properties.addAttribute("master_type_define_mark","#")
            END IF
            
            #單頭tree用local id
            LET ls_tmp = n.getAttribute("lid")
            CALL g_properties.addAttribute("master_field_lid",ls_tmp)

            #IF NOT cl_null(ls_tmp) THEN
            #   LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
            #   LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("master_var_title"), '.', ls_tmp, ','
            #END IF

            LET ls_tmp = g_properties.getValue("master_var_title"),".",n.getAttribute("lid")
            CALL g_properties.addAttribute("master_var_lid",ls_tmp)
            
            #單頭tree用parent pid
            IF g_properties.getValue("type_id") <> "q03" THEN
               LET ls_tmp = n.getAttribute("pid")
               CALL g_properties.addAttribute("master_field_pid",ls_tmp)
               LET ls_tmp = g_properties.getValue("master_var_title"),".",n.getAttribute("pid")
               CALL g_properties.addAttribute("master_var_pid",ls_tmp)
            ELSE
               LET ls_tmp = n.getAttribute("pid")
               CALL g_properties.addAttribute("master_field_lid",ls_tmp)
               LET ls_tmp = g_properties.getValue("master_var_title"),".",n.getAttribute("lid")
               CALL g_properties.addAttribute("master_var_lid",ls_tmp)
            END IF

            #單頭tree用type
            FOR li_idx = 1 TO 10
               LET ls_name = adzp152_create_name(li_idx, "master_type_mark", "<<<") 
               LET ls_value = "#"
               CALL g_properties.addAttribute(ls_name,ls_value)
            END FOR
            
            LET ls_tmp = ""
            FOR li_idx = 1 TO 7
            
               LET ls_name = adzp152_create_name(li_idx, "type", "<<<") 
               LET ls_value = n.getAttribute(ls_name)
               
               IF cl_null(ls_value) THEN
                  LET ls_name = "browser_max_lv"
                  LET ls_value = li_idx USING "<<<"
                  CALL g_properties.addAttribute(ls_name,ls_value)
                  LET ls_tmp = ls_tmp.subString(1,ls_tmp.getLength()-1)
                  CALL g_properties.addAttribute("browser_vars_type_list",ls_tmp)
                  EXIT FOR
               ELSE
                  LET ls_tmp = ls_tmp, "g_browser[li_idx].b_",ls_value,","
               END IF
               
               LET ls_name = adzp152_create_name(li_idx, "master_field_type", "<<<") 
               CALL g_properties.addAttribute(ls_name,ls_value)
               
               LET ls_value = g_properties.getValue("master_var_title"),".",ls_value
               LET ls_name = adzp152_create_name(li_idx, "master_var_type", "<<<") 
               CALL g_properties.addAttribute(ls_name,ls_value)
               
               LET ls_name = adzp152_create_name(li_idx, "master_type_mark", "<<<") 
               LET ls_value = ""
               CALL g_properties.addAttribute(ls_name,ls_value)
            END FOR
            
            #單頭pk
            LET ls_tmp = n.getAttribute("pk")
            CALL g_properties.addAttribute("master_field_pks",ls_tmp)   
            CALL adzp152_tab_master_primarykey(n.getAttribute("pk")) 
            IF NOT cl_null(ls_tmp) THEN
               LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
               LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
               WHILE lst_token.hasMoreTokens()
                  LET ls_token = lst_token.nextToken()
                  LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("master_var_title"), '.', ls_token, ','
               END WHILE
            END IF
            
            #單頭tree用描述欄位
            LET ls_tmp = n.getAttribute("desc")
            IF g_properties.getValue("type_id") = "q03" THEN
               CALL adzp152_create_desc(ls_tmp,'m')
            END IF

            #去掉最後的逗點 
            LET ls_field_allkeys = ls_field_allkeys.subString(1,ls_field_allkeys.getLength()-1)
            LET ls_var_allkeys = ls_var_allkeys.subString(1,ls_var_allkeys.getLength()-1)
            CALL g_properties.addAttribute("master_field_allkeys",ls_field_allkeys)   
            CALL g_properties.addAttribute("master_var_allkeys",ls_var_allkeys)   
            
            #因應i09,i10 寫入bs資訊
            #IF g_properties.getValue("type_id_t") = "q01" THEN
            #   CALL adzp152_create_bs_info(n.getAttribute("pk"))
            #END IF
            
            #speed(提速檔相關)
            LET ls_tmp = n.getAttribute("speed")            
            CALL g_properties.addAttribute("master_spped_tbl",ls_tmp)
            LET ls_tmp = adzp152_get_table_pre(ls_tmp)
            CALL g_properties.addAttribute("master_spped_tbl_pre",ls_tmp)
            CALL g_properties.addAttribute("master_spped_lid",n.getAttribute("slid"))
            CALL g_properties.addAttribute("master_spped_pid",n.getAttribute("spid"))
            CALL g_properties.addAttribute("master_spped_type",n.getAttribute("stype"))
            
            CALL adzp152_xml_dataset_search(n) 
            
         WHEN "body"
            CALL g_properties.addAttribute("detail_tbl_cnt",1)
            LET ls_field_allkeys = ""
            LET ls_var_allkeys = ""
            
            IF g_detail_chk THEN
               #多table資料擷取
               CALL adzp152_detail_multitable_info(n,1)
               LET n = n.getNext()
               CONTINUE WHILE
            ELSE
               #主table資料擷取
               CALL adzp152_detail_multitable_info(n,1)
            END IF
            
            LET g_detail_chk = TRUE
           
            #單身tree用帳別(type)
            LET ls_tmp = n.getAttribute("type")
            CALL g_properties.addAttribute("detail_field_type",ls_tmp)

            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_type",ls_tmp)
           
            #單身tree用local id
            LET ls_tmp = n.getAttribute("lid")
            CALL g_properties.addAttribute("detail_field_lid",ls_tmp)

            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_lid",ls_tmp)
            
            #單身tree用parent pid
            LET ls_tmp = n.getAttribute("pid")
            CALL g_properties.addAttribute("detail_field_pid",ls_tmp)

            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_pid",ls_tmp)

            #單身tree用描述欄位
            LET ls_tmp = n.getAttribute("desc")
            CALL g_properties.addAttribute("detail_field_desc",ls_tmp)
            
            #判斷desc是否為主table欄位
            IF ls_tmp.getIndexOf(g_properties.getValue("detail_tbl_prefix"),1) > 0 THEN
               CALL g_properties.addAttribute("detail_field_desc_chk",ls_tmp)
            ELSE
               CALL g_properties.addAttribute("detail_field_desc_chk","''")
            END IF
            
            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_desc",ls_tmp)
            
            #額外產生add-point段落
            LET ls_tmp = n.getAttribute("desc")
            IF g_properties.getValue("type_id") = "q03" THEN
              CALL adzp152_create_desc(ls_tmp,'m')
            END IF
            
            #對應到單頭的key
            CALL g_properties.addAttribute("detail_field_fks",n.getAttribute("fk"))   
            LET ls_tmp = n.getAttribute("fk")
            CALL adzp152_tab_detail_primarykey(ls_tmp,"fk")
            IF NOT cl_null(ls_tmp) THEN
               LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
               LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("master_var_allkeys"),","
            END IF
            
            #本身單身的key
            CALL g_properties.addAttribute("detail_field_pks",n.getAttribute("pk"))   
            LET ls_tmp = n.getAttribute("pk")
            CALL adzp152_tab_detail_primarykey(ls_tmp,"pk")
            IF NOT cl_null(ls_tmp) THEN
               LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
            END IF
            
            LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("detail_var_title"), '[l_ac].', ls_token, ','
            END WHILE
            
            #去掉最後的逗點
            LET ls_field_allkeys = ls_field_allkeys.subString(1,ls_field_allkeys.getLength()-1)
            LET ls_var_allkeys = ls_var_allkeys.subString(1,ls_var_allkeys.getLength()-1)
            CALL g_properties.addAttribute("detail_field_allkeys",ls_field_allkeys)
            CALL g_properties.addAttribute("detail_var_allkeys",ls_var_allkeys)
            CALL adzp152_xml_dataset_search(n)

            #page特殊處理(隱藏資料)
            CALL adzp152_page_mark()
            
            
         WHEN "sql"   #最底層,不會有子節點
            CASE n.getAttribute("id")
               WHEN "forupd_sql"
                  CALL g_properties.addAttribute("master_sql_forupd",n.getAttribute("query"))
               
               WHEN "forupd_sql_detail"
                  CALL g_properties.addAttribute("detail_sql_forupd",adzp152_field_filter(n.getAttribute("query"),"sql"))
               
               WHEN "b_fill_sql"
                  LET ls_tmp = adzp152_field_filter(n.getAttribute("query"),"sql")
                  CALL g_properties.addAttribute("detail_fill_sql",ls_tmp)
                  LET ls_tmp2 = ls_tmp.subString(1,ls_tmp.getIndexOf("WHERE",1)-1)
                  CALL g_properties.addAttribute("detail_fill_sql_pre",ls_tmp2)
                  LET ls_tmp2 = ls_tmp.subString(ls_tmp.getIndexOf("WHERE",1)-1,ls_tmp.getLength())
                  CALL g_properties.addAttribute("detail_fill_sql_post",ls_tmp2)
               
               WHEN "append"
                  #判斷是第幾個多語言table
                  LET ls_idx = g_properties.getValue("master_multi_table_idx")
                  IF cl_null(ls_idx) THEN
                     LET li_idx = 1
                  ELSE
                     LET li_idx = ls_idx + 1
                  END IF
                  CALL g_properties.addAttribute("master_multi_table_idx",li_idx)
                  CALL adzp152_master_multi_table(n)
                  CALL adzp152_master_multi_table_upd(n)
                  CALL adzp152_master_multi_table_bak(n)
                  CALL adzp152_update_item_action('m',n)

               WHEN "detail_append"
               #  CALL adzp152_detail_multi_table(n)
               #  CALL adzp152_update_item_action('d',n)

               OTHERWISE 
                  DISPLAY "ERROR(6): dataset內的",n.getAttribute("id"),"不符合需求定義"
            
            END CASE
            
         OTHERWISE 
            DISPLAY "ERROR(7): dataset內的",n.getTagName(),"節點不符合需求定義"
      END CASE 

      LET n = n.getNext()  #同一層 next sibling    
   END WHILE
   
   #先行取得第二階單身資料
   WHILE n2 IS NOT NULL
      IF n2.getTagName() = "body" THEN
         CALL adzp152_detail_multitable_info(n2,2)
      END IF
      LET n2 = n2.getNext()  #同一層 next sibling    
   END WHILE
   
END FUNCTION 


#+ 自定sql
PRIVATE FUNCTION adzp152_makesql(p_type)
   DEFINE p_type         STRING
   DEFINE l_sql          STRING
   DEFINE l_h_field_tmp  STRING
   DEFINE l_h_var_tmp    STRING
   DEFINE l_h_field      STRING
   DEFINE l_h_var        STRING
   DEFINE l_tmp          STRING
   DEFINE l_title        STRING
   DEFINE l_tok          base.StringTokenizer
   DEFINE l_mkey,l_vkey  STRING
   DEFINE li_tmp         INTEGER
   
   CASE p_type 
      WHEN "refresh"
         LET l_h_field_tmp = g_properties.getValue("master_fields_all")
         LET l_h_var_tmp = g_properties.getValue("master_vars_all")
       
         #過濾掉var_h_fill,master_fields_all中非單頭table的欄位
         LET l_tok = base.StringTokenizer.create(l_h_field_tmp,",")
         LET l_title = g_properties.getValue("title_master")
       
         WHILE l_tok.hasMoreTokens()
            LET l_tmp = l_tok.nextToken()
            IF adzp152_chk_intbl(g_properties.getValue("master_tbl_name"),l_tmp) THEN
               IF l_tmp.getIndexOf("(",1) THEN
                  LET l_tmp = l_tmp.subString(1,l_tmp.getIndexOf("(",1)-1)
               END IF
            IF cl_null(l_h_field) THEN
               LET l_h_field = l_tmp
            ELSE
               LET l_h_field = l_h_field, ",", l_tmp
            END IF
            END IF
         END WHILE
       
         LET l_tok = base.StringTokenizer.create(l_h_var_tmp,",")
         LET l_title = ".",g_properties.getValue("title_master")
         
         WHILE l_tok.hasMoreTokens()
            LET l_tmp = l_tok.nextToken()
            IF adzp152_chk_intbl(g_properties.getValue("master_tbl_name"),l_tmp) THEN
               IF cl_null(l_h_var) THEN
                  LET l_h_var = l_tmp
               ELSE
                  LET l_h_var = l_h_var, ",", l_tmp
               END IF
            END IF
         END WHILE
       
         IF cl_null(l_sql) THEN LET l_sql = "*" END IF
         LET l_sql = " SELECT UNIQUE ", l_h_field,"\n",
                     " INTO "  , l_h_var,"\n",
                     " FROM "  , g_properties.getValue("master_tbl_name"),"\n",
                     " WHERE " , g_properties.getValue("master_append_wc")

         #將所有的key併入搜尋條件
         LET li_tmp = "1"                         
         WHILE TRUE
            LET l_mkey = "master_field_pk",li_tmp USING "&&"
            LET l_vkey = "master_var_pk",li_tmp USING "&&"
            IF cl_null(g_properties.getValue(l_mkey)) THEN
               EXIT WHILE
            END IF
            LET l_sql = l_sql, g_properties.getValue(l_mkey)," = ",g_properties.getValue(l_vkey), " AND "
            LET li_tmp = li_tmp+1
         END WHILE
         
         LET l_sql = l_sql.subString(1,l_sql.getLength()-4)

   END CASE 
   
   RETURN l_sql 
   
END FUNCTION 


#+ convert fields to screen record #單身 查詢條件  cs_sr_name
PRIVATE FUNCTION adzp152_change_cs_sr_detail(p_name,p_list)  
   DEFINE p_name     STRING
   DEFINE p_list     STRING
   DEFINE l_prefix   STRING
   DEFINE lst_token  base.StringTokenizer
   DEFINE ls_token   STRING
   DEFINE l_new      STRING

   LET l_prefix = p_name
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
 
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()      
      LET l_new = l_new, l_prefix, "[1].", ls_token.trim(), ","      
   END WHILE   
   
   RETURN l_new.subString(1, l_new.getLength()-1)
      
END FUNCTION



#+ 定義g_browser 變數
PRIVATE FUNCTION adzp152_browser_fields_define(p_list,p_append_list,pk_list,ps_ref_tbl,ps_ref_wc,ps_ref_type)

   DEFINE p_list            STRING
   DEFINE p_append_list     STRING
   DEFINE pk_list           STRING
   DEFINE ps_ref_wc         STRING
   DEFINE ps_ref_tbl        STRING
   DEFINE ps_ref_type       STRING
   DEFINE l_prefix          STRING
   DEFINE lst_token         base.StringTokenizer
   DEFINE ls_token          STRING
   DEFINE lst_token2        base.StringTokenizer
   DEFINE ls_token2         STRING
   DEFINE l_new             STRING
   DEFINE l_str_token_count LIKE type_t.num5 
   DEFINE l_token_count     LIKE type_t.num5 
   DEFINE li_diff_tbl       LIKE type_t.num5  #辨識是否有兩種以上table
   DEFINE ls_tabid          STRING
   DEFINE ls_tmp            STRING
   DEFINE ls_tmp2           STRING
   DEFINE ls_ref_type       STRING
   DEFINE ls_ref_wc         STRING
   DEFINE ls_replace        LIKE type_t.num5
   DEFINE ls_type           STRING
   DEFINE ls_tmp_tbl        STRING
   DEFINE ls_name           STRING

   #處理必要變動欄位
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_new = ""
   LET ls_tmp_tbl = ""   
   LET li_diff_tbl = FALSE
   LET l_str_token_count = lst_token.countTokens()
      
   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1

      #修正table代碼
      IF ls_token.getIndexOf('(',1) = 0 THEN
         LET l_prefix = adzp152_define_table_name(ls_token)
      END IF
      
      #i05 insert 用
      LET ls_tmp2 = ls_tmp2,"?"

      #比對是否存在不同table
      IF NOT li_diff_tbl AND ls_tabid IS NOT NULL THEN
         IF NOT ls_tabid.equals(l_prefix) THEN
            LET li_diff_tbl = TRUE
         END IF
      END IF
      LET ls_tabid = l_prefix
     
      #非一般參照欄位
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_type = ls_token.subString(ls_token.getIndexOf('(',1)+1,ls_token.getIndexOf(')',1)-1)
         LET ls_name = "b_",ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
         CASE ls_type.toUpperCase()
            WHEN "STRING"
               #LET ls_fb_new = ls_fb_new, ls_name, " STRING" #STRING過於複雜, 導致讀資料時發生錯誤, 暫時停用
               LET l_new = l_new, ls_name, " LIKE type_t.chr500"
            WHEN "DATETIME"
               LET l_new = l_new, ls_name, " DATETIME YEAR TO SECOND"
            WHEN "TIMESTAMP"
               LET l_new = l_new, ls_name, " DATETIME YEAR TO FRACTION(5)"
            WHEN "BLOB"
               LET l_new = l_new, ls_name, " LIKE type_t.", ls_type
               LET ls_tmp = g_properties.getValue("general_lob")
               LET ls_tmp = ls_tmp, "   LOCATE g_browser.",ls_name," IN FILE ", "\n"
               CALL g_properties.addAttribute("general_lob_start",ls_tmp)
               LET ls_tmp = g_properties.getValue("general_lob_end")
               LET ls_tmp = ls_tmp, "   FREE g_browser.",ls_name, "\n"
               CALL g_properties.addAttribute("general_lob_end",ls_tmp)
            WHEN "CLOB"
               LET l_new = l_new, ls_name, " LIKE type_t.", ls_type
               LET ls_tmp = g_properties.getValue("general_lob")
               LET ls_tmp = ls_tmp, "   LOCATE g_browser.",ls_name," IN FILE ", "\n"
               CALL g_properties.addAttribute("general_lob_start",ls_tmp)
               LET ls_tmp = g_properties.getValue("general_lob_end")
               LET ls_tmp = ls_tmp, "   FREE g_browser.",ls_name, "\n"
               CALL g_properties.addAttribute("general_lob_end",ls_tmp)
            OTHERWISE
               IF ls_type.getIndexOf('.',1) > 0 THEN
                  LET l_new = l_new, ls_name, " LIKE ", ls_type
               ELSE
                  LET l_new = l_new, ls_name, " LIKE type_t.", ls_type
               END IF
         END CASE
         #LET l_new = l_new, li_space SPACES, ls_name , " LIKE type_t.", ls_type
         LET ls_tmp_tbl = ls_tmp_tbl,li_space SPACES,ls_token.subString(1,ls_token.getIndexOf('(',1)-1), " VARCHAR(500),\n"
      ELSE
         LET l_new = l_new, li_space SPACES, "b_",ls_token.trim() , " LIKE ", l_prefix ,ls_token.trim()
         LET ls_tmp_tbl = ls_tmp_tbl, li_space SPACES, ls_token.trim() , " VARCHAR(500),\n"
      END IF 
      
      IF l_str_token_count != l_token_count THEN
         LET l_new = l_new ,",\n",li_space SPACES
         LET ls_tmp2 = ls_tmp2,","
      END IF 
   END WHILE
   LET ls_tmp = l_new.subString( 1, l_new.getLength())
   LET ls_tmp_tbl = ls_tmp_tbl.subString( 1, ls_tmp_tbl.getLength()-2)
   CALL g_properties.addAttribute("browser_fields_define",ls_tmp)
   CALL g_properties.addAttribute("browser_tmp_tbl_define", ls_tmp_tbl)
   CALL g_properties.addAttribute("browser_fields_qm", ls_tmp2)
   
   #設定Browser內若有使用referance時的wc
   IF li_diff_tbl THEN
      IF NOT cl_null(ps_ref_wc) AND NOT cl_null(ps_ref_tbl) THEN
         CASE ps_ref_type
            WHEN "left"
               LET ls_ref_type = " LEFT JOIN "
            WHEN "right"
               LET ls_ref_type = " RIGHT JOIN " 
            WHEN "inner"
               LET ls_ref_type = " INNER JOIN "
            OTHERWISE
               LET ls_ref_type = " LEFT JOIN "
         END CASE   

         LET ls_tmp = ""

         LET lst_token = base.StringTokenizer.create(ps_ref_wc, ',')
         LET lst_token2 = base.StringTokenizer.create(ps_ref_tbl, ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_token2 = lst_token2.nextToken()
            
            LET ls_tmp = ls_tmp, 
                         ls_ref_type,ls_token2, 
                         " ON ", ls_token, " "
            
         END WHILE
         
         LET ls_tmp = adzp152_replace_var(ls_tmp)

         CALL g_properties.addAttribute("browser_addiction_sql",ls_tmp)
         
      #ELSE
      #   DISPLAY "ERROR(8):Browser內設定Referance表,但未設定關聯"
      END IF
      
   END IF
   
   #檢視是否有值
   IF NOT cl_null(ls_tmp) THEN
      RETURN TRUE
   ELSE
      RETURN FALSE
   END IF
   
END FUNCTION


#+ convert fields to var
PRIVATE FUNCTION adzp152_master_fields_define(pi_page,p_list)  
   DEFINE p_list                  STRING
   DEFINE l_prefix                STRING
   DEFINE lst_token               base.StringTokenizer
   DEFINE ls_token                STRING
   DEFINE li_i                    LIKE type_t.num10    #空幾格
   DEFINE pi_page                 LIKE type_t.num5
   DEFINE l_token_count           LIKE type_t.num5
   DEFINE l_str_token_count       LIKE type_t.num5
   DEFINE ls_master_fields_define STRING       #form_head_field代換變數
   DEFINE ls_var_h_fill           STRING       #var_h_fill代換變數
   DEFINE ls_master_fields_all    STRING       #var_h_fill代換變數
   DEFINE ls_fb_new               STRING       #form_head_field填入值
   DEFINE ls_vb_new               STRING       #var_h_fill填入值
   DEFINE ls_tmp                  STRING
   DEFINE ls_tmp2                 STRING
   DEFINE ls_value                STRING
   DEFINE ls_type                 STRING       #變數型態
   DEFINE ls_name                 STRING       #變數名稱
   DEFINE ls_master_all_var       STRING       #所有單頭變數
   DEFINE ls_master_upd_var       STRING       #主table的單頭變數
   DEFINE ls_field_sql_head       STRING       #所有必需insert進單頭table的欄位 
   DEFINE ls_name_list            STRING
   DEFINE ls_gzza019              LIKE gzza_t.gzza019
   DEFINE ls_gzza020              LIKE gzza_t.gzza020
   DEFINE ls_master_field         STRING
   DEFINE ls_parameter_list       STRING
   DEFINE ls_j                    LIKE type_t.num5
   DEFINE ls_str                  STRING
   DEFINE ls_prog_name            LIKE gzza_t.gzza001
   DEFINE ls_var_field            STRING


   #定義處理變數名稱
   LET ls_master_fields_define = "master_fields_define"
   LET ls_var_h_fill = "master_vars_all"
   LET ls_master_fields_all = "master_fields_all"
   
   #master_fields_all過濾&整理 
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf("(",1) THEN
         LET ls_vb_new = ls_vb_new,ls_token.subString(1,ls_token.getIndexOf("(",1)-1),","
      ELSE
         LET ls_vb_new = ls_vb_new,ls_token,","
      END IF
   END WHILE
   
   #去除最後逗點
   LET ls_vb_new = ls_vb_new.subString(1,ls_vb_new.getLength())

   CALL g_properties.addAttribute(ls_master_fields_all, ls_vb_new)

   LET ls_vb_new = ""
   
   LET li_i = 1
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_str_token_count = lst_token.countTokens()

   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1
 
      #逐個串接DEFINE for form_head_master
      LET ls_fb_new = ls_fb_new, (li_space*li_i) SPACES

      #逐個串接 for var_h_fill
      IF pi_page = 1 THEN 
         LET ls_vb_new = ls_vb_new, g_properties.getValue("master_var_title")
      ELSE
         LET ls_tmp = "master_var_title",pi_page USING "<<<"
         LET ls_vb_new = ls_vb_new, g_properties.getValue(ls_tmp)
      END IF

      #如果標示欄位名稱的字串中含有 (),表示要使用 type_t, 並且是用()內指定的型態
      CASE 
         WHEN ls_token.getIndexOf("(",1) 
            LET l_prefix = "type_t."
            LET ls_name = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            LET ls_master_field = ls_name.trim(),","
            LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf(")",1)-1)
            CASE ls_type.toUpperCase()
               WHEN "STRING"
                  #LET ls_fb_new = ls_fb_new, ls_name, " STRING" #STRING過於複雜, 導致讀資料時發生錯誤, 暫時停用
                  LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.chr500"
               WHEN "DATETIME"
                  LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO SECOND"
               WHEN "TIMESTAMP"
                  LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO FRACTION(5)"
               WHEN "BLOB"
                  LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               WHEN "CLOB"
                  LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               OTHERWISE
                  ##190826-00045#1 -S-
                  #LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                  IF ls_type.getIndexOf(".",1) > 0 THEN
                     LET ls_fb_new = ls_fb_new, ls_name, " LIKE ", ls_type
                  ELSE
                     LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                  END IF
                  ##190826-00045#1 -E-
            END CASE 
            LET ls_vb_new = ls_vb_new, ".",ls_token.subString(1,ls_token.getIndexOf("(",1)-1)

            IF cl_null(ls_name_list) THEN
               LET ls_name_list = ls_name
            ELSE
               LET ls_name_list = ls_name_list,",",ls_name
            END IF
         WHEN ls_token.getIndexOf("crtdt",1) OR ls_token.getIndexOf("moddt",1) OR ls_token.getIndexOf("cnfdt",1)
            LET ls_fb_new = ls_fb_new, ls_token, " DATETIME YEAR TO SECOND"
            LET ls_vb_new = ls_vb_new, ".",ls_token.trim()

            IF cl_null(ls_name_list) THEN
               LET ls_name_list = ls_token
            ELSE
               LET ls_name_list = ls_name_list,",",ls_token
            END IF
         OTHERWISE
            LET l_prefix = adzp152_define_table_name(ls_token)
            LET ls_fb_new = ls_fb_new, ls_token.trim(), " LIKE ", l_prefix, ls_token.trim()
            LET ls_vb_new = ls_vb_new, ".",ls_token.trim()

            IF cl_null(ls_name_list) THEN
               LET ls_name_list = ls_token
            ELSE
               LET ls_name_list = ls_name_list,",",ls_token
            END IF
      END CASE
      IF l_str_token_count != l_token_count THEN
         LET ls_fb_new = ls_fb_new, ", \n"
         LET ls_vb_new = ls_vb_new, ","
      END IF
   END WHILE

   IF pi_page = 1 THEN 
   ELSE
      LET ls_master_fields_define = ls_master_fields_define,pi_page USING "<<<<"
      LET ls_var_h_fill = ls_var_h_fill,pi_page USING "<<<<"
   END IF 

   CALL g_properties.addAttribute(ls_master_fields_define, ls_fb_new.trim())
   CALL g_properties.addAttribute(ls_var_h_fill, ls_vb_new.trim())
   
   #額外處理var_sql_head,field_sql_head
   LET ls_tmp = g_properties.getValue("master_vars_all")
   LET lst_token = base.StringTokenizer.create(ls_tmp, ',')

   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF adzp152_chk_intbl(g_properties.getValue("master_tbl_name"),ls_token) THEN
        LET ls_master_upd_var = ls_master_upd_var, ls_token,","
      END IF
      LET ls_master_all_var = ls_master_all_var, ls_token,","
   END WHILE
   LET ls_master_upd_var = ls_master_upd_var.subString(1,ls_master_upd_var.getLength()-1)
   LET ls_master_all_var = ls_master_all_var.subString(1,ls_master_all_var.getLength()-1)
   
   CALL g_properties.addAttribute("master_vars_all",ls_master_all_var)  
   CALL g_properties.addAttribute("master_vars_update",ls_master_upd_var)   

   #取得azzi900中設定的Q類程式要串查的主維護程式資訊
   LET ls_prog_name = g_properties.getValue("app_id")
   SELECT gzza019,gzza020 INTO ls_gzza019,ls_gzza020 FROM gzza_t
    WHERE gzza001 = ls_prog_name

   IF NOT cl_null(ls_gzza020) THEN
      LET ls_parameter_list = ls_gzza020
      #把空白取代掉
      LET ls_parameter_list = cl_replace_str(ls_parameter_list," ","")

      LET lst_token = base.StringTokenizer.create(ls_parameter_list,",")
      LET ls_j = 0
      LET ls_name_list = ls_name_list.trim(),","
      WHILE lst_token.hasMoreTokens()
         LET ls_j = ls_j + 1
         LET ls_master_field = lst_token.nextToken(),","
         IF ls_name_list.getIndexOf(ls_master_field,1) THEN
            LET ls_var_field = g_properties.getValue("master_var_title"),".",ls_master_field.subString(1,ls_master_field.getLength()-1)
            LET ls_str = "master_maintain_prog_parameter_",ls_j USING "<<<"
            CALL g_properties.addAttribute(ls_str,ls_var_field)
         END IF
      END WHILE
      IF ls_j > 0 THEN
         CALL g_properties.addAttribute("master_maintain_prog_parameter_cnt",ls_j)
      END IF
   END IF

   IF NOT cl_null(ls_gzza019) THEN
      LET ls_str = "master_maintain_prog"
      CALL g_properties.addAttribute(ls_str,ls_gzza019)
   END IF

   LET ls_tmp = g_properties.getValue("master_fields_all")
   LET lst_token = base.StringTokenizer.create(ls_tmp, ',')

   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF adzp152_chk_intbl(g_properties.getValue("master_tbl_name"),ls_token) THEN
         LET ls_field_sql_head = ls_field_sql_head, ls_token,","
      END IF
   END WHILE
   LET ls_field_sql_head = ls_field_sql_head.subString(1,ls_field_sql_head.getLength()-1)
   #DISPLAY "upd_head_fields:",ls_field_sql_head
   CALL g_properties.addAttribute("master_fields_update",ls_field_sql_head)

   #複製段清空key欄位desc
   LET ls_tmp = ""
   LET lst_token = base.StringTokenizer.create(p_list, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      #先判斷是否為參照欄位
      IF ls_token.getIndexOf('_desc',1) > 0 THEN
         LET ls_tmp = ls_token.subString(1,ls_token.getIndexOf('_desc',1)-1)
         #再判斷是否為key欄位的參照欄位
         LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
         IF ls_tmp2.getIndexOf(ls_tmp,1) > 0 THEN
            LET ls_value = ls_value, 
                           "   LET ", g_properties.getValue("master_var_title"),".",
                           ls_token.subString(1,ls_token.getIndexOf('(',1)-1), " = ''\n",
                           "   DISPLAY BY NAME ", g_properties.getValue("master_var_title"),".",
                           ls_token.subString(1,ls_token.getIndexOf('(',1)-1), "\n"
         END IF
      END IF
   END WHILE
    
   CALL g_properties.addAttribute("general_repro_reset",ls_value)   
   
END FUNCTION


#+ convert fields to var
PRIVATE FUNCTION adzp152_detail_fields_define(pi_page,p_list)  
   DEFINE p_list                   STRING
   DEFINE l_prefix                 STRING
   DEFINE lst_token                base.StringTokenizer
   DEFINE ls_token                 STRING
   DEFINE li_i                     LIKE type_t.num10    
   DEFINE pi_page                  LIKE type_t.num5
   DEFINE l_token_count            LIKE type_t.num5
   DEFINE l_str_token_count        LIKE type_t.num5
   DEFINE ls_detail_fields_define  STRING               #form_body_field代換變數
   DEFINE ls_detail_vars_all       STRING               #detail_vars_all代換變數
   DEFINE ls_detail_fields_all     STRING               #detail_fields_all代換變數            
   DEFINE ls_fb_new                STRING               #form_body_field填入值
   DEFINE ls_vb_new                STRING               #detail_vars_all填入值
   DEFINE ls_tmp                   STRING
   DEFINE ls_name                  STRING 
   DEFINE ls_type                  STRING    
   DEFINE ls_page                  STRING
   DEFINE ls_str                   STRING             #add by joyce 14/03/14
   DEFINE ls_form_field            STRING             #add by joyce 14/03/14
   DEFINE ls_field_list            STRING             #YSC-E40007

   #定義處理變數名稱
   LET ls_detail_fields_define = "detail_fields_define"
   LET ls_detail_vars_all = "detail_vars_all"
   LET ls_detail_fields_all = "detail_fields_all"
   
   LET ls_page = pi_page USING "<<<"
   
   CALL g_properties.addAttribute("detail_page_field"||ls_page,p_list)

   LET li_i = 1
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_str_token_count = lst_token.countTokens()

   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1
      
      #取得每個page的第一個欄位
      IF l_token_count = 1 THEN
         LET ls_tmp = g_properties.getValue("detail_var_title"||ls_page),"[l_ac].",ls_token
         CALL g_properties.addAttribute("detail_fisrt_var"||ls_page,ls_tmp)
         CALL g_properties.addAttribute("detail_fisrt_field"||ls_page,ls_token)
      END IF
 
      #逐個串接DEFINE for form_body_detail
      LET ls_fb_new = ls_fb_new, (li_space*li_i) SPACES

      #逐個串接 for detail_vars_all
      IF pi_page = 1 THEN 
         LET ls_vb_new = ls_vb_new, g_properties.getValue("detail_var_title")
      ELSE
         LET ls_tmp = "detail_var_title",pi_page USING "<<<"
         LET ls_vb_new = ls_vb_new, g_properties.getValue(ls_tmp)
      END IF

      #如果標示欄位名稱的字串中含有 (),表示要使用 type_t, 並且是用()內指定的型態
      CASE
         WHEN ls_token.getIndexOf("(",1)
            LET l_prefix = "type_t."
            LET ls_name = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf(")",1)-1)

            IF ls_type.getIndexOf('.',1) > 0 THEN
               LET ls_fb_new = ls_fb_new, ls_name, " LIKE ", ls_type
               LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            ELSE
               CASE ls_type.toUpperCase()
                  WHEN "STRING"
                     #LET ls_fb_new = ls_fb_new, ls_name, " STRING"
                  #  LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.chr500"   #YSC-E40007 mark

                     #YSC-E40007 ---start---
                     # 若有做串查的欄位，欄位型態應維持STRING，以避免超連結字串太長被截掉
                     LET ls_tmp = adzp152_create_name(pi_page, "detail_hyper_field_list", "<<<")
                     LET ls_field_list = g_properties.getValue(ls_tmp)
                     IF ls_field_list.getIndexOf(ls_name,1) > 0 THEN
                        LET ls_fb_new = ls_fb_new, ls_name, " STRING"
                     ELSE
                        LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.chr500"
                     END IF
                     #YSC-E40007 --- end ---
                  WHEN "DATETIME"
                     LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO SECOND"
                  WHEN "TIMESTAMP"
                     LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO FRACTION(5)"
                  OTHERWISE
                     ##190826-00045#1 -S-
                     #LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                     IF ls_type.getIndexOf(".",1) > 0 THEN
                        LET ls_fb_new = ls_fb_new, ls_name, " LIKE ", ls_type
                     ELSE
                        LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                     END IF
                     ##190826-00045#1 -E-
               END CASE 
               LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            END IF

            # YSC-E70001 ---start---
            # sel欄位改為可變動設定
            IF ls_name = "sel" THEN
               CALL g_properties.addAttribute("detail_sel_set","Y")
            END IF
            # YSC-E70001 --- end ---

         WHEN ls_token.getIndexOf("crtdt",1) OR ls_token.getIndexOf("moddt",1) OR ls_token.getIndexOf("cnfdt",1)
            LET ls_fb_new = ls_fb_new, ls_token, " DATETIME YEAR TO SECOND"
            LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.trim()
            LET ls_name = ls_token
            
         OTHERWISE
            #修正table代碼
            LET l_prefix = adzp152_define_table_name(ls_token)
            LET ls_fb_new = ls_fb_new, ls_token.trim(), " LIKE ", l_prefix ,ls_token.trim()
            LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.trim()
            LET ls_name = ls_token
            
      END CASE

      IF l_str_token_count != l_token_count THEN
         LET ls_fb_new = ls_fb_new, ", \n"
         LET ls_vb_new = ls_vb_new, ","
      END IF

   END WHILE

   IF pi_page = 1 THEN 
   ELSE
      LET ls_detail_fields_define = ls_detail_fields_define,pi_page USING "<<<<"
      LET ls_detail_vars_all = ls_detail_vars_all,pi_page USING "<<<<"
      #LET ls_vb_new = ",",ls_vb_new
      LET ls_detail_fields_all = ls_detail_fields_all,pi_page USING "<<<<"
   END IF 

   CALL g_properties.addAttribute(ls_detail_fields_define, ls_fb_new.trim())
   CALL g_properties.addAttribute(ls_detail_vars_all, ls_vb_new.trim())
   CALL g_properties.addAttribute(ls_detail_fields_all, p_list.trim())
   
   # YSC-E40001 ---start---
   # 因為目前q03與q01共用q01樣板，但仍需做部分微調
   IF g_properties.getValue("type_id") = "q01" THEN
      IF g_properties.getValue("type_id_t") = "q01" THEN   # tab檔紀錄的樣板代號
         # root_array的定義只有在q03樣板才有
         LET ls_str = " "
         CALL g_properties.addAttribute("detail_root_array_define",ls_str)
      ELSE
         # root_array的定義只有在q03樣板才有
         LET ls_str = "DEFINE g_browser_root  DYNAMIC ARRAY OF INTEGER   #root資料所在"
         CALL g_properties.addAttribute("detail_root_array_define",ls_str)
      END IF
   END IF
   # YSC-E40001 --- end ---
END FUNCTION


#+ 設定單頭key值
PRIVATE FUNCTION adzp152_tab_master_primarykey(p_keys)

   DEFINE p_prefix   STRING
   DEFINE p_keys     STRING
   DEFINE lst_token  base.StringTokenizer
   DEFINE ls_token   STRING
   DEFINE ls_name    STRING
   DEFINE ls_value   STRING
   DEFINE li_cnt     LIKE type_t.num5 

   LET li_cnt = 1   
   LET lst_token = base.StringTokenizer.create(p_keys.trim(), ',')
   
   WHILE lst_token.hasMoreTokens()
       LET ls_token = lst_token.nextToken()
       LET ls_name = "master_field_pk", li_cnt USING "&&"
       CALL g_properties.addAttribute(ls_name,ls_token) 
       LET ls_value = g_properties.getValue("master_var_title"),".",ls_token
       LET ls_name = "master_var_pk", li_cnt USING "&&"
       CALL g_properties.addAttribute(ls_name,ls_value)

       LET li_cnt =  li_cnt + 1
       
   END WHILE 
   LET li_cnt =  li_cnt - 1

   CALL g_properties.addAttribute("master_pk_num",li_cnt )

END FUNCTION



#+ 設定單身key值
PRIVATE FUNCTION adzp152_tab_detail_primarykey(p_keys,p_type)

   DEFINE p_prefix   STRING
   DEFINE p_type     STRING
   DEFINE p_keys     STRING
   DEFINE lst_token  base.StringTokenizer
   DEFINE ls_token   STRING
   DEFINE ls_name    STRING
   DEFINE ls_value   STRING
   DEFINE li_cnt     LIKE type_t.num5

   LET li_cnt = 1   
   LET p_prefix = "detail_field_",p_type
   LET lst_token = base.StringTokenizer.create(p_keys.trim(), ',')
   
   WHILE lst_token.hasMoreTokens()
       LET ls_token = lst_token.nextToken()
       
       LET ls_name = p_prefix, li_cnt USING "&&"
       CALL g_properties.addAttribute(ls_name,ls_token) 
       IF p_type = "fk" THEN
          LET ls_name = "master_field_pk", li_cnt USING "&&"
          LET ls_value = g_properties.getValue("master_var_title"),".",g_properties.getValue(ls_name)
          LET ls_name = "detail_var_fk", li_cnt USING "&&"
          CALL g_properties.addAttribute(ls_name,ls_value)
       ELSE
          LET ls_value = g_properties.getValue("detail_var_title"),"[l_ac].",ls_token
          LET ls_name = "detail_var_pk", li_cnt USING "&&"
          CALL g_properties.addAttribute(ls_name,ls_value)
       END IF

       LET li_cnt =  li_cnt + 1
       
   END WHILE 

   LET li_cnt =  li_cnt - 1
   
   CALL g_properties.addAttribute("detail_"||p_type||"_num",li_cnt )

END FUNCTION



#+ 設定upd_body_fields及cs_body_fields 組單身欄位頁簽
PRIVATE FUNCTION adzp152_set_upd_body_fields()

   DEFINE l_k                    LIKE type_t.num5 
   DEFINE l_upd_body_vars        STRING
   DEFINE l_upd_body_fields      STRING
   DEFINE l_upd_body_vars_page   STRING 
   DEFINE l_upd_body_fields_page STRING
   DEFINE l_tok                  base.StringTokenizer
   DEFINE l_sr                   STRING 
   DEFINE l_detail               STRING
   DEFINE l_table                STRING 
   DEFINE l_table2               STRING 
   DEFINE l_str_token_count      LIKE type_t.num5 
   DEFINE l_token_count          LIKE type_t.num5  
   DEFINE l_tag_id               STRING 
   DEFINE l_page                 LIKE type_t.num5  
   DEFINE l_str_page             STRING
   DEFINE l_upd_defaults         STRING
   DEFINE l_tmp                  STRING
   DEFINE l_detail_page          STRING
   DEFINE l_upd_body_fields_tmp  STRING #temp table用
   
   LET l_detail = g_properties.getValue("detail_tbl_prefix")
   LET l_detail = l_detail CLIPPED
   LET l_detail = "g_",l_detail,"_d"

   LET l_page = g_properties.getValue("page")  

   FOR l_k = 1 TO l_page
      LET l_upd_body_vars = ""  
      LET l_upd_body_fields = ""

      #第一頁簽的處理是特別的
      IF l_k = 1 THEN
         LET l_str_page = ""
         LET l_detail_page = l_detail.subString(1,l_detail.getIndexOf("_",3)-1),l_str_page,"_d"
         LET l_tag_id = "detail_fields_all" 
         LET l_table = g_properties.getValue("detail_tbl_name") 
         LET l_table2 = l_table
         LET l_table = l_table.subString(1,l_table.getIndexOf("_",1)-1)
      ELSE 
         LET l_str_page = l_k USING "<<<"
         #取 g_xxx_d 的第一碼,到第二個底線前一碼的意思
         LET l_detail_page = l_detail.subString(1,l_detail.getIndexOf("_",3)-1),l_str_page,"_d"
         LET l_tag_id = "detail_fields_all",l_str_page 
         LET l_table = g_properties.getValue("detail_tbl_name"||l_str_page)
         LET l_table2 = l_table         
         LET l_table = l_table.subString(1,l_table.getIndexOf("_",1)-1)
      END IF   
      
     
      LET l_upd_body_vars_page = "detail_vars_update",l_str_page
      LET l_upd_body_fields_page = "detail_fields_update",l_str_page

      LET l_tok = base.StringTokenizer.create(g_properties.getValue(l_tag_id),",")
      LET l_str_token_count = l_tok.countTokens()
      LET l_token_count = 0

      WHILE l_tok.hasMoreTokens()
         LET l_sr = l_tok.nextToken()
         LET l_token_count = l_token_count + 1
         LET l_tmp = g_properties.getValue("detail_field_pks")
         IF adzp152_chk_intbl(l_table2,l_sr) THEN
            IF l_tmp.getIndexOf(l_sr,1) > 0 AND l_k > 1 THEN
            ELSE
               LET l_upd_body_vars = l_upd_body_vars,l_detail_page,"[l_ac].",l_sr,","
               LET l_upd_body_fields = l_upd_body_fields,l_sr,","
               LET l_upd_body_fields_tmp = l_upd_body_fields_tmp, l_sr, " VARCHAR(500),\n"
            END IF
         END IF
      END WHILE 
      
      IF l_upd_body_vars.subString(l_upd_body_vars.getLength(),l_upd_body_vars.getLength()) = "," THEN
         LET l_upd_body_vars = l_upd_body_vars.subString(1,l_upd_body_vars.getLength()-1)
      END IF
      
      IF l_upd_body_fields.subString(l_upd_body_fields.getLength(),l_upd_body_fields.getLength()) = "," THEN
         LET l_upd_body_fields = l_upd_body_fields.subString(1,l_upd_body_fields.getLength()-1)
      END IF

      LET l_upd_body_fields_tmp = l_upd_body_fields_tmp.subString(1,l_upd_body_fields_tmp.getLength()-2)
      CALL g_properties.addAttribute("detail_tmp_tbl_define",l_upd_body_fields_tmp)
      CALL g_properties.addAttribute(l_upd_body_vars_page,l_upd_body_vars)
      CALL g_properties.addAttribute(l_upd_body_fields_page,l_upd_body_fields)

   END FOR   
   
END FUNCTION

#+ 判定該節點以下的子節點屬於何種類型並進行處理
PUBLIC FUNCTION adzp152_parse_element(p_node,pi_lv)
   DEFINE p_node       om.DomNode
   DEFINE pi_lv        LIKE type_t.num10     #退後 pi_lv 個層級
   DEFINE l_node       om.DomNode
   DEFINE ls_type      STRING
   DEFINE ls_ref_type  STRING
   DEFINE l_n          LIKE type_t.num10 
   DEFINE l_return     STRING 
   DEFINE ls_tmp       STRING
   DEFINE li_mchk      INTEGER
   DEFINE li_mchk2     INTEGER
   
   LET li_mchk = 0
   LET li_mchk2 = 0

   #若該節點沒有子節點則不做任何處理
   IF p_node.getChildCount() > 0 THEN 
      FOR l_n = 1 TO p_node.getChildCount()
         LET l_node = p_node.getChildByIndex(l_n)
         LET ls_type = l_node.getTagName()
    
         #根據不同的元件呼叫不同的製造Function
         CASE ls_type
            WHEN "reference"
               LET ls_ref_type = l_node.getAttribute("type")
               IF ls_ref_type = "lang" THEN
                  LET ls_ref_type = adzp152_create_reference(l_node,pi_lv)
               ELSE
                  LET l_return = l_return,adzp152_create_reference(l_node,pi_lv)
               END IF
            
            WHEN "mreference"
               LET li_mchk = li_mchk + 1
               LET l_return = l_return,adzp152_create_mreferance(l_node,pi_lv,li_mchk)
            
            WHEN "mcode"
               LET li_mchk2 = li_mchk2 + 1
               LET l_return = l_return,adzp152_create_mcode(l_node,pi_lv,li_mchk2)
            
         END CASE 
         
      END FOR 
      
   END IF 
   
   RETURN l_return
   
END FUNCTION 


#+ 生成controlp元件(common欄位自動產生)
PRIVATE FUNCTION adzp152_create_controlp_common(ps_field,pi_lv)
   DEFINE ps_field        STRING 
   DEFINE pi_lv           LIKE type_t.num10 
   DEFINE ls_tbl          STRING
   DEFINE ls_field        STRING
   DEFINE ls_var          STRING
   DEFINE ls_return       STRING 
   
   #取得table名稱
   IF g_properties.getValue("location") = "head" THEN
      LET ls_tbl = g_properties.getValue("master_tbl_name")
   ELSE
      LET ls_tbl = g_properties.getValue("detail_tbl_name")
   END IF
   
   #取得欄位名稱
   LET ls_field = ps_field 

   #先組合變數名稱
   IF g_properties.getValue("location") = "head" THEN
      LET ls_var = g_properties.getValue("master_var_title"),".",ls_field
   ELSE
      LET ls_var = g_properties.getValue("detail_var_title"),"[l_ac].",ls_field
   END IF
   
   #格式 CALL q_gzxa('table名稱','欄位名稱',TRUE, FALSE, 變數名稱) RETURNING 變數名稱
   LET ls_return = (pi_lv*li_space) SPACES,
                  "CALL q_common('", ls_tbl,"','", ls_field,"',TRUE,FALSE,",ls_var,") RETURNING ", ls_var,"\n",
                  (pi_lv*li_space) SPACES,
                  "DISPLAY BY NAME ", ls_var, "\n",
                  (pi_lv*li_space) SPACES,
                  "NEXT FIELD ", ls_field
                  
   RETURN ls_return
   
END FUNCTION 


#+ 判斷該欄位參照的table名稱
PUBLIC FUNCTION adzp152_define_table_name(ps_field)
   DEFINE ps_field     STRING
   DEFINE ls_field     CHAR(100)
   DEFINE ls_table     CHAR(100)
   DEFINE ls_return    STRING
   DEFINE ls_prefix    CHAR(100)
   DEFINE li_cnt       LIKE type_t.num5   #200210-00032
   
   #取table前四碼
   LET ls_prefix = '%_t'
   LET ls_field  = ps_field
   
   #200210-00032 ---modify start---
#  #改由gztz進行比對
#  LET ls_table = ""
#  SELECT gztz001 INTO ls_table FROM gztz_t 
#   WHERE gztz002 = ls_field AND gztz001 LIKE ls_prefix
#   
#  IF cl_null(ls_table) THEN
#     DISPLAY "ERROR(10):欄位",ls_field CLIPPED, "不存在於資料庫中, 請重新檢查!"
#  END IF

   #先比對資料是否重複
   LET ls_table = ""
   SELECT COUNT(*) INTO li_cnt FROM dzeb_t
    WHERE dzeb002 = ls_field #AND dzeb001 LIKE ls_prefix

   IF li_cnt > 1 THEN
      DISPLAY "ERROR(10-1):欄位",ls_field CLIPPED, "重複存在於資料庫(dzeb_t)中, 請到gztz中檢查該欄位出現多次的原因!"
      EXIT PROGRAM
   END IF

   #改由gztz進行比對
   LET ls_table = ""
   SELECT dzeb001 INTO ls_table FROM dzeb_t
    WHERE dzeb002 = ls_field #AND dzeb001 LIKE ls_prefix

   IF cl_null(ls_table) THEN
      DISPLAY "ERROR(10-2):欄位",ls_field CLIPPED, "不存在於資料庫(dzeb_t)中, 請重新檢查!"
   END IF
   #200210-00032 --- modify end ---

   LET ls_return = ls_table CLIPPED, '.'

   RETURN ls_return
   
END FUNCTION


#+ 判斷該欄位參照的table名稱
PUBLIC FUNCTION adzp152_get_table_pre(ps_tbl_name)
   DEFINE ps_tbl_name STRING
   DEFINE ls_tbl_pre  STRING
   DEFINE li_idx      INTEGER
   
   LET li_idx = ps_tbl_name.getIndexOf("_",1) - 1
   
   LET ls_tbl_pre = ps_tbl_name.subString(1,li_idx)
   
   RETURN ls_tbl_pre

END FUNCTION


#+ 判斷該欄位參照的table名稱
PRIVATE FUNCTION adzp152_get_table_name(pn_root)
   DEFINE pn_root     om.DomNode
   DEFINE lnl_data    om.NodeList
   DEFINE ln_node     om.DomNode
   DEFINE ln_hnode    om.DomNode
   DEFINE ln_bnode    om.DomNode
   DEFINE ls_hfile    STRING
   DEFINE ls_bfile    STRING
   DEFINE ls_key_list STRING
   DEFINE ls_str      STRING             #YSC-E30002
   DEFINE ls_exist    LIKE type_t.chr1   #YSC-E30002

   LET lnl_data = pn_root.selectByTagName("dataset")
   LET ln_node = lnl_data.item(1)
 
   #此處無法使用CASE WHEN,其原因是大多數的樣板兩種都要做,非唯一執行
   #有單頭的樣板
   IF adzp152_type_decide("m") THEN
      LET ls_exist = 'Y'     #YSC-E30002
      #取出單頭table
      LET lnl_data = ln_node.selectByTagName("head")
      LET ln_hnode = lnl_data.item(1)
      IF ln_hnode IS NULL THEN 
         # YSC-E30002 ---modify start---
         # 目前只有q04樣板會有單頭，但q04樣板不一定會有設dataset資訊
         DISPLAY "WARNING(16):",g_properties.getValue("type_id"),"樣板未設定單頭相關資料!"
      #  DISPLAY "ERROR(16):",g_properties.getValue("type_id"),"樣板必須設定單頭相關資料, 程式終止!"
      #  EXIT PROGRAM
         LET ls_exist = 'N'
         # YSC-E30002 --- modify end ---
      END IF
      LET ls_hfile = ln_hnode.getAttribute("id")
      CALL g_properties.addAttribute("master_tbl_name",ls_hfile)
      LET ls_hfile = adzp152_get_table_pre(ls_hfile)
      CALL g_properties.addAttribute("master_tbl_prefix",ls_hfile)
      LET ls_key_list = ln_hnode.getAttribute("pk")
      CALL g_properties.addAttribute("master_field_allkeys",ls_key_list)

      LET ls_str = "master_dataset_",g_properties.getValue("type_id"),"_exist"   #YSC-E30002
      CALL g_properties.addAttribute(ls_str, ls_exist)   #YSC-E30002
   END IF
 
   #有單身的樣板
   IF adzp152_type_decide("d") THEN
      LET ls_exist = 'Y'     #YSC-E30002
      #取出單身table
      LET lnl_data = ln_node.selectByTagName("body")
      LET ln_bnode = lnl_data.item(1)
      IF ln_bnode IS NULL THEN
         # 目前只有q02樣板一定會有dataset的資訊，其餘樣板可能有可能沒有
         #YSC-E30002 ---modify start---
         IF g_properties.getValue("type_id") = "q02" THEN
            DISPLAY "ERROR(17):",g_properties.getValue("type_id"),"樣板必須設定單身相關資料, 程式終止!"
            EXIT PROGRAM
         ELSE
            DISPLAY "WARNING(17):",g_properties.getValue("type_id"),"樣板未設定單身相關資料!"
            LET ls_exist = 'N'
         END IF
         #YSC-E30002 --- modify end ---
      END IF
      LET ls_bfile = ln_bnode.getAttribute("id")
      CALL g_properties.addAttribute("detail_tbl_name",ls_bfile)
      
      LET ls_bfile = adzp152_get_table_pre(ls_bfile)
      CALL g_properties.addAttribute("detail_tbl_prefix",ls_bfile)

      LET ls_str = "detail_dataset_",g_properties.getValue("type_id"),"_exist"   #YSC-E30002
      CALL g_properties.addAttribute(ls_str, ls_exist)   #YSC-E30002
      
   END IF

END FUNCTION

#150529-00031 ---start---
#+ 取得dataset中各單頭單身的資訊
PUBLIC FUNCTION adzp152_get_dataset_table_info(ls_n)
   DEFINE ls_n                om.DomNode
   DEFINE ln_n                om.DomNode
   DEFINE lt_n                om.DomNode
   DEFINE li_cnt              LIKE type_t.num5
   DEFINE li_page             STRING
   DEFINE li_str              STRING
   DEFINE li_field            STRING
   DEFINE li_n_f              LIKE type_t.num5   #第一階單身個數
   DEFINE li_n_o              LIKE type_t.num5   #其他單身個數
   DEFINE li_detail           STRING
   DEFINE lnl_data            om.NodeList
   DEFINE ln_bnode            om.DomNode


   LET li_n_f = 0
   LET li_n_o = 0
   #取得第一個子節點 returns the first child node
   LET ln_n = ls_n.getFirstChild()
   WHILE ln_n IS NOT NULL
      CASE ln_n.getTagName()
         WHEN "dataset"   #取 structure->section
            FOR li_cnt = 1 TO ln_n.getChildCount()
               LET lt_n = ln_n.getChildByIndex(li_cnt)
               LET lnl_data = lt_n.selectByTagName("body")
               LET ln_bnode = lnl_data.item(1)
               IF ln_bnode IS NOT NULL THEN
                  #確認是第一階單身還是其他單身
                  LET li_detail = NULL
                  LET li_detail = ln_bnode.getAttribute("detail")
                  IF cl_null(li_detail) THEN
                     LET li_n_f = li_n_f + 1
                     LET li_str = adzp152_create_name(li_n_f, "detail_first_page", "<<<")
                  ELSE
                     LET li_n_o = li_n_o + 1
                     LET li_str = adzp152_create_name(li_n_o, "detail_other_page", "<<<")
                  END IF

                  #將此table包含的頁籤資訊寫入記錄下來
                  LET li_page = ln_bnode.getAttribute("page")
                  CALL g_properties.addAttribute(li_str,li_page)
               END IF
            END FOR

            #紀錄第一階單身或其他單身的筆數
            LET li_str = "detail_first_page_num"
            CALL g_properties.addAttribute(li_str,li_n_f)
            LET li_str = "detail_other_page_num"
            CALL g_properties.addAttribute(li_str,li_n_o)

         OTHERWISE
            CALL adzp152_get_dataset_table_info(ln_n)
      END CASE
      LET ln_n = ln_n.getNext()  #同一層 next sibling

   END WHILE
END FUNCTION
#150529-00031 --- end ---

#YSC-E40007 --- start---
#+ 取得各PAGE有做串查功能的欄位清單
PRIVATE FUNCTION adzp152_get_hyper_fields(ls_n)
   DEFINE ls_n                om.DomNode
   DEFINE ln_n                om.DomNode
   DEFINE lt_n                om.DomNode
   DEFINE li_cnt              LIKE type_t.num5
   DEFINE li_page             LIKE type_t.num5
   DEFINE li_str              STRING
   DEFINE li_field            STRING
   

   #取得第一個子節點 returns the first child node
   LET ln_n = ls_n.getFirstChild()
   WHILE ln_n IS NOT NULL
      CASE ln_n.getTagName()                                 
         WHEN "section"   #取 structure->section
            CASE 
               WHEN ln_n.getAttribute("id") = "detail_show"
                  LET li_page = ln_n.getAttribute("page")
                  LET li_field = NULL

                  FOR li_cnt = 1 TO ln_n.getChildCount()
                     LET lt_n = ln_n.getChildByIndex(li_cnt)
                     IF lt_n.getTagName() = "cluster" THEN
                        LET li_str = adzp152_create_name(li_page, "detail_hyper_field_list", "<<<")
                        LET li_field = g_properties.getValue(li_str)
                        IF NOT cl_null(li_field) THEN
                           LET li_field = li_field,","
                        END IF
                        LET li_field = li_field,lt_n.getAttribute("id")
                        CALL g_properties.addAttribute(li_str,li_field)
                     END IF
                  END FOR
            END CASE 

         OTHERWISE
            CALL adzp152_get_hyper_fields(ln_n)
      END CASE  
      LET ln_n = ln_n.getNext()  #同一層 next sibling 
      
   END WHILE 
END FUNCTION
#YSC-E40007 --- end ---

#+ 判斷樣板是否需進行此項判斷
PRIVATE FUNCTION adzp152_type_decide(ps_type)
   DEFINE ps_type STRING            
   # m ( master  -   單頭   )
   # d ( detail  -   單身   )
   # t ( tree    -   樹狀   )
   # b ( brwoser - 瀏覽頁籤 )

   CASE ps_type
      #YSC-E30002 ---modify start---
      WHEN 'm' #有單頭
         IF g_properties.getValue("type_id") = "q04" THEN
            RETURN TRUE
         END IF
         
      WHEN 'd' #有單身
         IF g_properties.getValue("type_id") = "q01"  OR
            g_properties.getValue("type_id") = "q02"  OR
            g_properties.getValue("type_id") = "q03"  OR
            g_properties.getValue("type_id") = "q04"  THEN
            RETURN TRUE
         END IF
         
      WHEN 't' #有TREE
         IF g_properties.getValue("type_id") = "q03" THEN
            RETURN TRUE
         END IF
         
   #  WHEN 'b' #有BROWSER
   #     IF g_properties.getValue("type_id") = "q01" THEN
   #        RETURN TRUE
   #     END IF    

   #  WHEN 'm' #有單頭
   #     IF g_properties.getValue("type_id") = "q01"  OR
   #        g_properties.getValue("type_id") = "q03"  THEN
   #        RETURN TRUE
   #     END IF
   #     
   #  WHEN 'd' #有單身
   #     IF g_properties.getValue("type_id") = "q02"  OR
   #        g_properties.getValue("type_id") = "q03"  THEN
   #        RETURN TRUE
   #     END IF
   #     
   #  WHEN 't' #有TREE
   #     IF g_properties.getValue("type_id") = "q03" THEN
   #        RETURN TRUE
   #     END IF
   #     
   #  WHEN 'b' #有BROWSER
   #     IF g_properties.getValue("type_id") = "q01" THEN
   #        RETURN TRUE
   #     END IF    
      #YSC-E30002 --- modify end ---
   END CASE

   RETURN FALSE
    
END FUNCTION


#+ 取代掉$var並還原成正確字串
PUBLIC FUNCTION adzp152_replace_var(ps_str)
DEFINE ps_str      STRING
DEFINE ls_str      STRING
DEFINE ls_replace  LIKE type_t.num10

   LET ls_str = ps_str
   #進行資料比對並取代
   WHILE TRUE
      LET ls_replace= ls_str.getIndexOf("$var",1)
      
      #若有$var則開始取代
      IF ls_replace > 0 THEN
         LET ls_str = ls_str.subString(1,ls_replace-1),"'\"||",ls_str.subString(ls_replace+4,ls_str.getLength())
         LET ls_replace= ls_str.getIndexOf("$var",1)
         #$var為成對存在, 若只有單一存在則顯示錯誤訊息
         IF ls_replace = 0 THEN
            DISPLAY "ERROR(13):",ps_str,"定義錯誤, 請檢查該參數內容!"
         ELSE
            #進行重組並取代
            LET ls_str = ls_str.subString(1,ls_replace-1),"||\"'",ls_str.subString(ls_replace+4,ls_str.getLength())
         END IF
      ELSE         
         EXIT WHILE
      END IF
   END WHILE
   
   RETURN ls_str

END FUNCTION

#enterprise code/site code各項處理
FUNCTION adzp152_append_field()
   DEFINE l_tmp                 STRING
   DEFINE lc_tmp                CHAR(500)
   DEFINE l_master_table        STRING
   DEFINE l_master_field_ent    STRING
   DEFINE l_master_field_site   STRING
   DEFINE l_detail_table        STRING
   DEFINE l_detail_field_ent    STRING
   DEFINE l_detail_field_site   STRING
   DEFINE l_master_ent          BOOLEAN
   DEFINE l_detail_ent          BOOLEAN
   DEFINE l_master_site         BOOLEAN
   DEFINE l_detail_site         BOOLEAN
   DEFINE l_master_wc           STRING
   DEFINE l_detail_wc           STRING
   DEFINE l_master_cl           STRING
   DEFINE l_detail_cl           STRING
   DEFINE l_master_define       STRING
   DEFINE l_detail_define       STRING
   DEFINE l_master_wc_s         STRING
   DEFINE l_detail_wc_s         STRING
   DEFINE l_master_var          STRING
   DEFINE l_detail_var          STRING
   DEFINE l_master_field        STRING
   DEFINE l_detail_field        STRING
   DEFINE l_chk_site            BOOLEAN
   DEFINE l_chk_site_pk         STRING
   DEFINE lc_chk_site_pk        CHAR(500)

   #判斷單頭table是否有enterprise
   LET l_tmp = g_properties.getValue("master_tbl_name")
   LET l_master_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_master"),"ent"
   LET l_master_field_ent = l_tmp.toLowerCase() 
   IF cl_getField(l_master_table CLIPPED, l_master_field_ent CLIPPED) THEN
      CALL g_properties.addAttribute("master_field_ent",l_master_field_ent)
      LET l_master_ent = TRUE
   ELSE
      LET l_master_ent = FALSE
   END IF
   
   #判斷單身table是否有enterprise
   LET l_tmp = g_properties.getValue("detail_tbl_name")
   LET l_detail_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_detail"),"ent"
   LET l_detail_field_ent = l_tmp.toLowerCase() 
   IF cl_getField(l_detail_table CLIPPED, l_detail_field_ent CLIPPED) THEN
      CALL g_properties.addAttribute("detail_field_ent",l_detail_field_ent)
      LET l_detail_ent = TRUE
   ELSE
      LET l_detail_ent = FALSE
   END IF
   
   #判斷單頭table是否有site
   LET l_tmp = g_properties.getValue("master_tbl_name")
   LET l_master_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_master"),"site"
   LET l_master_field_site = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("master_field_allkeys")
   
   #判定單頭table是否有site
   IF l_tmp.getIndexOf(l_master_field_site,1) = 0 THEN
      CALL g_properties.addAttribute("master_field_site",l_master_field_site)
      LET l_chk_site = TRUE
   ELSE
      LET l_chk_site = FALSE
   END IF
   
   #判定單頭site是否為key
   LET lc_tmp = g_properties.getValue("master_tbl_name")
   SELECT dzed004 INTO lc_chk_site_pk FROM dzed_t
    WHERE dzed001 = lc_tmp
      AND dzed003 = 'P'
   
   LET l_chk_site_pk = lc_chk_site_pk
   IF cl_getField(l_master_table CLIPPED, l_master_field_site CLIPPED ) AND 
      l_chk_site AND
      l_chk_site_pk.getIndexOf(l_master_field_site,1) > 0 THEN
      LET l_master_site = TRUE
   ELSE
      LET l_master_site = FALSE
   END IF
   
   #判斷單身table是否有site
   LET l_tmp = g_properties.getValue("detail_tbl_name")
   LET l_detail_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_detail"),"site"
   LET l_detail_field_site = l_tmp.toLowerCase() 


   #判定單身site是否為key
   LET lc_tmp = g_properties.getValue("detail_tbl_name")
   SELECT dzed004 INTO lc_chk_site_pk FROM dzed_t
    WHERE dzed001 = lc_tmp
      AND dzed003 = 'P'

   LET l_chk_site_pk = lc_chk_site_pk
   IF cl_getField(l_detail_table CLIPPED, l_detail_field_site CLIPPED) AND 
      l_chk_site_pk.getIndexOf(l_detail_field_site,1) > 0 AND
      l_chk_site THEN
   #  CALL g_properties.addAttribute("detail_field_site",l_detail_field_site)
      LET l_detail_site = TRUE
   ELSE
      LET l_detail_site = FALSE
   END IF

   #若單頭有enterprise欄位, 進行所需資料組合
   IF l_master_ent THEN
      LET l_master_wc_s  = l_master_wc_s  , l_master_field_ent CLIPPED, " = '\" ||g_enterprise|| \"' AND "
      LET l_master_wc    = l_master_wc    , l_master_field_ent CLIPPED, " = g_enterprise AND "
      LET l_master_field = l_master_field , l_master_field_ent CLIPPED, ", "
      LET l_master_var   = l_master_var   , "g_enterprise, "
      LET l_master_define = 
          l_master_define, "LET ",g_properties.getValue("master_var_title"),".",l_master_field_ent, " = g_enterprise \n" 
      LET l_master_cl = l_master_cl, l_master_field_ent, " = ? AND "
   END IF
   
   #若單身有enterprise欄位, 進行所需資料組合
   IF l_detail_ent THEN
      LET l_detail_wc_s  = l_detail_wc_s  , l_detail_field_ent CLIPPED, " = '\" ||g_enterprise|| \"' AND "
      LET l_detail_wc    = l_detail_wc    , l_detail_field_ent CLIPPED, " = g_enterprise AND "
      LET l_detail_field = l_detail_field , l_detail_field_ent CLIPPED, ", "
      LET l_detail_var   = l_detail_var   , "g_enterprise, "
      LET l_detail_define = 
          l_detail_define, "LET ",g_properties.getValue("detail_var_title"),".",l_detail_field_ent, " = g_enterprise \n" 
      LET l_detail_cl = l_detail_cl, l_detail_field_ent, " = ? AND "
   END IF
      
   #若單頭有site欄位, 進行所需資料組合
   IF l_master_site THEN
      LET l_master_wc_s  = l_master_wc_s  , l_master_field_site CLIPPED, " = '\" ||g_site|| \"' AND "
      LET l_master_wc    = l_master_wc    , l_master_field_site CLIPPED, " = g_site AND "
      LET l_master_field = l_master_field , l_master_field_site CLIPPED, ", "
      LET l_master_var   = l_master_var   , "g_site, "
      LET l_master_define = 
          l_master_define, "LET ",g_properties.getValue("master_var_title"),".",l_master_field_site, " = g_site \n" 
      LET l_master_cl = l_master_cl, l_master_field_site, " = ? AND "
   END IF        
     
   #若單身有site欄位, 進行所需資料組合
   IF l_detail_site THEN
      LET l_detail_wc_s  = l_detail_wc_s  , l_detail_field_site CLIPPED, " = '\" ||g_site|| \"' AND "
      LET l_detail_wc    = l_detail_wc    , l_detail_field_site CLIPPED, " = g_site AND "
      LET l_detail_field = l_detail_field , l_detail_field_site CLIPPED, ", "
      LET l_detail_var   = l_detail_var   , "g_site, "
      LET l_detail_define = 
          l_detail_define, "LET ",g_properties.getValue("detail_var_title"),".",l_detail_field_site, " = g_site \n" 
      LET l_detail_cl = l_detail_cl, l_detail_field_site, " = ? AND "
   END IF

   #單頭段
   CALL g_properties.addAttribute("master_append_wc_s"  , l_master_wc_s  )
   CALL g_properties.addAttribute("master_append_wc"    , l_master_wc    )
   CALL g_properties.addAttribute("master_field_append" , l_master_field )
   CALL g_properties.addAttribute("master_var_append"   , l_master_var   )
   #CALL g_properties.addAttribute("master_define_append", l_master_define)
                           
   #單身段
   CALL g_properties.addAttribute("detail_append_wc_s"  , l_detail_wc_s  )
   CALL g_properties.addAttribute("detail_append_wc"    , l_detail_wc    )
   CALL g_properties.addAttribute("detail_field_append" , l_detail_field )
   CALL g_properties.addAttribute("detail_var_append"   , l_detail_var   )
   
END FUNCTION       

#特定樣板進行註解特定段落處理
FUNCTION adzp152_mark()
   DEFINE l_mdl    CHAR(10)
   DEFINE l_tmp    STRING
   DEFINE li_idx   INTEGER

   LET l_mdl = g_properties.getValue("type_id_t")

   CASE

      WHEN l_mdl = "q02" OR l_mdl = "q01"
         #若有ent欄位
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            LET l_tmp = "EXECUTE b_fill_cnt_pre USING g_enterprise, g_site INTO g_tot_cnt"
            CALL g_properties.addAttribute("detail_cnt_using_ent",l_tmp )
         END IF
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            LET l_tmp = "EXECUTE b_fill_cnt_pre USING g_enterprise INTO g_tot_cnt"
            CALL g_properties.addAttribute("detail_cnt_using_ent",l_tmp )
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            LET l_tmp = "EXECUTE b_fill_cnt_pre USING g_site INTO g_tot_cnt"
            CALL g_properties.addAttribute("detail_cnt_using_ent",l_tmp )
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            LET l_tmp = "EXECUTE b_fill_cnt_pre INTO g_tot_cnt"
            CALL g_properties.addAttribute("detail_cnt_using_ent",l_tmp )
         END IF
         
      OTHERWISE
      
   END CASE
   
END FUNCTION

#+ Tree架構desc段落處理
FUNCTION adzp152_create_desc(ps_list,ps_type)
   DEFINE ps_list      STRING
   DEFINE ps_type      STRING
   DEFINE ls_name      STRING
   DEFINE lst_token    base.StringTokenizer
   DEFINE ls_addpoint  STRING
   DEFINE ls_tmp1      STRING
   DEFINE ls_tmp2      STRING
   DEFINE ls_token     STRING
   
   CALL g_properties.addAttribute("browser_field_desc",ps_list)
   
   #判斷所用之變數名稱
   IF ps_type = 'm' THEN
      IF cl_null(ps_list) THEN
         LET ps_list = g_properties.getValue("master_field_pk01")
      END IF
      LET lst_token = base.StringTokenizer.create(ps_list, ',')
      
      #分析欄位清單並處理
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #LET ls_tmp1 = ls_tmp1,"      LET g_browser[l_ac].b_",ls_token," = g_browser[l_ac].b_",ls_token," CLIPPED\n"
         LET ls_tmp2 = ls_tmp2,"           g_browser[l_ac].b_",ls_token,","
         IF lst_token.hasMoreTokens() THEN
            LET ls_tmp2 = ls_tmp2, "'-',\n" 
         END IF
      END WHILE
      
      #進行組合
      LET ls_addpoint = ls_tmp1,
                        "      LET g_browser[l_ac].b_show = \n",
                        ls_tmp2,"\n",
                        "             '(',g_browser[l_ac].b_",g_properties.getValue("master_field_lid"),",')'\n"
                        
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_name  = "desc_show.show"
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = ls_addpoint
   #  CALL adzp152_insert_addpoint(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint)        #YSC-E50004 mark
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))   #YSC-E50004 modify     
      
   ELSE
      IF cl_null(ps_list) THEN
         LET ps_list = g_properties.getValue("master_field_pk01")
      END IF
      LET lst_token = base.StringTokenizer.create(ps_list, ',')
      
      #分析欄位清單並處理
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #LET ls_tmp1 = ls_tmp1,"      LET g_browser[l_ac].b_",ls_token," = g_browser[l_ac].b_",ls_token," CLIPPED\n"
         LET ls_tmp2 = ls_tmp2,"           g_browser[l_ac].b_",ls_token,","
         IF lst_token.hasMoreTokens() THEN
            LET ls_tmp2 = ls_tmp2, "'-',\n" 
         END IF
      END WHILE
      
      #進行組合
      LET ls_addpoint = ls_tmp1,
                        "      LET g_browser[l_ac].b_show = \n",
                        ls_tmp2,"\n",
                        "             '(',g_browser[l_ac].b_",g_properties.getValue("detail_field_lid"),",')'\n"
                        
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_name  = "desc_show.show"
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = ls_addpoint
   #  CALL adzp152_insert_addpoint(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint)        #YSC-E50004 mark
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))   #YSC-E50004 modify     
      
   END IF

END FUNCTION


#+ 產生虛擬的bs info
FUNCTION adzp152_create_bs_info(ps_key_fields)
   DEFINE ps_key_fields   STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_bsvars       STRING
   DEFINE ls_key_fields   STRING
   
   CALL g_properties.addAttribute("bs_order", ps_key_fields )

   #去除特別欄位參照內容
   LET lst_token = base.StringTokenizer.create(ps_key_fields, ',')
   LET ls_tmp = '' 
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf('(',1) > 0 OR 
         ls_token.getIndexOf(g_properties.getValue("master_tbl_prefix"),1) = 0 THEN
         LET ls_tmp = ls_tmp, "'',"
      ELSE
         LET ls_tmp = ls_tmp, ls_token, ','
      END IF
   END WHILE
   LET ls_tmp = ls_tmp.substring(1,ls_tmp.getLength()-1)
   CALL g_properties.addAttribute("browser_fields_all", ls_tmp )
   
   #組合所有browser變數
   LET lst_token = base.StringTokenizer.create(ps_key_fields, ',')
   LET ls_bsvars = ""
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_token = ls_token.subString(1, ls_token.getIndexOf('(',1)-1)
      END IF      
      LET ls_bsvars = ls_bsvars,"g_browser[g_cnt].b_",ls_token
      IF lst_token.hasMoreTokens() THEN
         LET ls_bsvars = ls_bsvars,","
      END IF
   END WHILE               
   CALL g_properties.addAttribute("browser_vars_all", ls_bsvars )
   
   #ps_key_fields加工 
   LET ls_tmp = ""
   LET lst_token = base.StringTokenizer.create(g_properties.getValue("master_fields_define"), ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf('DATETIME YEAR TO SECOND',1) THEN
         LET ls_tmp = ls_tmp, ls_token
      END IF
   END WHILE
   
   LET lst_token = base.StringTokenizer.create(ps_key_fields, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_tmp.getIndexOf(ls_token,1) THEN
         LET ls_key_fields = ls_key_fields, ls_token, "(datetime)"
      ELSE
         LET ls_key_fields = ls_key_fields, ls_token
      END IF
      IF lst_token.hasMoreTokens() THEN
         LET ls_key_fields = ls_key_fields, ','
      END IF
   END WHILE
   
   #組合browser欄位定義
   IF NOT adzp152_browser_fields_define(ls_key_fields,"",ls_key_fields,"","","") THEN
      #DISPLAY "ERROR(5):設定g_browse變數發生問題"
   END IF
            
END FUNCTION 

#+ 確定該欄位是否於該table中
PUBLIC FUNCTION adzp152_chk_intbl(ps_table,ps_field)
   DEFINE ps_table      STRING
   DEFINE ps_field      STRING
   DEFINE ls_table      STRING
   DEFINE ls_field      STRING
   DEFINE ls_table_pre  STRING
   DEFINE ls_length     INTEGER
   
   LET ls_table = ps_table
   
   LET ls_field = ps_field

   #變數而非欄位
   IF ls_field.getIndexOf(".",1) > 0 THEN
      LET ls_field = ls_field.subString(ls_field.getIndexOf(".",1)+1,ls_field.getLength())
   END IF
   
   #若有含括號
   IF ls_field.getIndexOf("(",1) > 0 THEN
      LET ls_field = ls_field.subString(1,ls_field.getIndexOf("(",1)-1)
   END IF
   
   IF cl_getField(ls_table,ls_field) THEN
      RETURN TRUE
   ELSE
      RETURN FALSE
   END IF
   

END FUNCTION

#+ 查找欄位所在的page(名稱)
PUBLIC FUNCTION adzp152_find_page(ps_field,ps_pages)
   DEFINE ps_field        STRING
   DEFINE ps_pages        STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_name         STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_return       STRING
   DEFINE lb_chk          LIKE type_t.num5   #YSC-E70003


   LET lb_chk = TRUE   #YSC-E70003
   FOR li_idx = 1 TO g_properties.getValue("page")
      IF NOT cl_null(ps_pages) THEN
         LET ls_tmp = li_idx USING "<<<"
         IF ps_pages.getIndexOf(ls_tmp,1) = 0 THEN
            CONTINUE FOR
         END IF
      END IF
      LET ls_name = adzp152_create_name(li_idx, "detail_fields_define", "<<<") 
      LET ls_tmp = g_properties.getValue(ls_name)
      
      #確定尋找的欄位在單身的哪個page
      IF ls_tmp.getIndexOf(ps_field,1) > 0 THEN
         LET ls_tmp = li_idx USING "<<<"
         LET lb_chk = FALSE   #YSC-E70003
         EXIT FOR
      END IF
   END FOR
   LET ls_tmp = adzp152_create_name(ls_tmp, "detail_var_title", "<<<") 
   LET ls_return = g_properties.getValue(ls_tmp),"[l_ac]"

   #YSC-E70003 ---start---
   IF lb_chk THEN
      LET ls_return = ""
   END IF
   #YSC-E70003 --- end ---

   RETURN ls_return

END FUNCTION

#150529-00031 ---start---
#+ 查找欄位所在的page(名稱)
PUBLIC FUNCTION adzp152_find_table_page(ps_field,ps_pages)
   DEFINE ps_field        STRING
   DEFINE ps_pages        STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_name         STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_return       STRING
   DEFINE lb_chk          LIKE type_t.num5   #YSC-E70003
   DEFINE ls_table_page   STRING
   DEFINE ls_str          STRING
   DEFINE ls_i            LIKE type_t.num5
   DEFINE ls_array_idx    STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_n            LIKE type_t.num5


   LET lb_chk = TRUE   #YSC-E70003
   FOR li_idx = 1 TO g_properties.getValue("page")
      IF NOT cl_null(ps_pages) THEN
         LET ls_tmp = li_idx USING "<<<"
         IF ps_pages.getIndexOf(ls_tmp,1) = 0 THEN
            CONTINUE FOR
         END IF
      END IF
      LET ls_name = adzp152_create_name(li_idx, "detail_fields_define", "<<<")
      LET ls_tmp = g_properties.getValue(ls_name)

      #確定尋找的欄位在單身的哪個page
      IF ls_tmp.getIndexOf(ps_field,1) > 0 THEN
         #確認欄位是否在第一階單身取出的
         LET ls_n = g_properties.getValue("detail_first_page_num")
         IF ls_n > 0 THEN
            FOR ls_i = 1 TO ls_n
               #先取出dataset中定義的該table是取出哪幾個page的資料
               LET ls_str = adzp152_create_name(ls_i, "detail_first_page", "<<<")
               LET ls_table_page = g_properties.getValue(ls_str)
               IF NOT cl_null(ls_table_page) THEN
                  #因為一個table有可能會包含多個page的資料，所以須再做判斷
                  LET lst_token = base.StringTokenizer.create(ls_table_page, ',')
                  LET ls_array_idx = NULL
                  WHILE lst_token.hasMoreTokens()
                     LET ls_token = lst_token.nextToken()

                     #比對目前所在的page是否有在此table的範圍內
                     IF li_idx = ls_token THEN
                        LET ls_array_idx = "g_detail_idx"
                        EXIT WHILE
                     END IF
                  END WHILE
                  IF NOT cl_null(ls_array_idx) THEN
                     EXIT FOR
                  END IF
               END IF
            END FOR
         END IF

         IF cl_null(ls_array_idx) THEN
            #確認欄位是否在其他單身取出的
            LET ls_n = g_properties.getValue("detail_other_page_num")
            IF ls_n > 0 THEN
               FOR ls_i = 1 TO ls_n
                  #先取出dataset中定義的該table是取出哪幾個page的資料
                  LET ls_str = adzp152_create_name(ls_i, "detail_other_page", "<<<")
                  LET ls_table_page = g_properties.getValue(ls_str)
                  IF NOT cl_null(ls_table_page) THEN
                     #因為一個table有可能會包含多個page的資料，所以須再做判斷
                     LET lst_token = base.StringTokenizer.create(ls_table_page, ',')
                     LET ls_array_idx = NULL
                     WHILE lst_token.hasMoreTokens()
                        LET ls_token = lst_token.nextToken()

                        #比對目前所在的page是否有在此table的範圍內
                        IF li_idx = ls_token THEN
                           LET ls_array_idx = "g_detail_idx2"
                           EXIT WHILE
                        END IF
                     END WHILE
                     IF NOT cl_null(ls_array_idx) THEN
                        EXIT FOR
                     END IF
                  END IF
               END FOR
            END IF
         END IF

         LET ls_tmp = li_idx USING "<<<"
         LET lb_chk = FALSE   #YSC-E70003
         EXIT FOR
      END IF
   END FOR


   LET ls_tmp = adzp152_create_name(ls_tmp, "detail_var_title", "<<<")
   LET ls_return = g_properties.getValue(ls_tmp),"[",ls_array_idx,"]"

   #YSC-E70003 ---start---
   IF lb_chk THEN
      LET ls_return = ""
   END IF
   #YSC-E70003 --- end ---

   RETURN ls_return

END FUNCTION
#150529-00031 --- end ---

#+ 查找欄位所在的page(數字)
PUBLIC FUNCTION adzp152_find_page_num(ps_field,ps_pages)
   DEFINE ps_field        STRING
   DEFINE ps_pages        STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_name         STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_return       STRING
   
   FOR li_idx = 1 TO g_properties.getValue("page")
      IF NOT cl_null(ps_pages) THEN
         LET ls_tmp = li_idx USING "<<<"
         IF ps_pages.getIndexOf(ls_tmp,1) = 0 THEN
            CONTINUE FOR
         END IF
      END IF 
      LET ls_name = adzp152_create_name(li_idx, "detail_fields_define", "<<<") 
      LET ls_tmp = g_properties.getValue(ls_name)

      #確定尋找的欄位在單身的哪個page
      IF ls_tmp.getIndexOf(ps_field,1) > 0 THEN
         LET ls_tmp = li_idx USING "<<<"
         EXIT FOR
      END IF
   END FOR
   LET ls_return = ls_tmp

   RETURN ls_return

END FUNCTION 

#+ 產生browser filter相關段落
PUBLIC FUNCTION adzp152_create_borwser_filter(ps_list)
   DEFINE ps_list                  STRING
   DEFINE lst_token                base.StringTokenizer
   DEFINE ls_token                 STRING
   DEFINE ls_filter_field          STRING
   DEFINE ls_field_filter_bs       STRING
   DEFINE ls_display_info          STRING
   DEFINE ls_display_condition     STRING
   DEFINE ls_field                 STRING
   
   IF NOT adzp152_type_decide("b") THEN
      RETURN
   END IF
   
   LET lst_token = base.StringTokenizer.create(ps_list, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_field = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
DISPLAY "ls_field:",ls_field
      #非主要table欄位
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_field = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
         IF NOT cl_getField(g_properties.getValue("master_tbl_name"),ls_field) THEN
            CONTINUE WHILE
         ELSE
            LET ls_token = ls_field
         END IF
      END IF
      
      LET ls_filter_field = ls_filter_field, ls_token, ','
      
      LET ls_field_filter_bs = ls_field_filter_bs, 's_browse[1].b_', ls_token, ','
      
      LET ls_display_info = ls_display_info, 
                            12 SPACES, "DISPLAY ${general_prefix}_filter_parser('",ls_token,"')",
                            " TO s_browse[1].b_", ls_token, "\n"
                            
      LET ls_display_condition = ls_display_condition, 
                                 3 SPACES, "CALL ${general_prefix}_filter_show('",ls_token,"')", "\n"
   END WHILE

   LET ls_filter_field = ls_filter_field.subString(1,ls_filter_field.getLength()-1)
   LET ls_field_filter_bs = ls_field_filter_bs.subString(1,ls_field_filter_bs.getLength()-1)
   LET ls_display_info = ls_display_info.subString(1,ls_display_info.getLength()-1)
   LET ls_display_condition = ls_display_condition.subString(1,ls_display_condition.getLength()-1)

   CALL g_properties.addAttribute("browser_fields_filter",ls_filter_field)
   CALL g_properties.addAttribute("browser_fields_filter_bs",ls_field_filter_bs)
   CALL g_properties.addAttribute("browser_display_info",ls_display_info)
   CALL g_properties.addAttribute("browser_display_condition",ls_display_condition)
DISPLAY "ls_filter_field:",ls_filter_field
DISPLAY "ls_field_filter_bs:",ls_field_filter_bs
END FUNCTION 


#+ 產生slice相關段落(特定樣板使用)
PRIVATE FUNCTION adzp152_create_slice()
   DEFINE ls_tmp            STRING
   DEFINE li_tmp            INTEGER
   DEFINE l_master_dataset  STRING   #YSC-E30002
   DEFINE l_detail_dataset  STRING   #YSC-E30002
   
   #產出main段落
   LET ls_tmp = g_properties.getValue("type_id_t")
   #200804-00006 ---start---
   #Q類不需要這些SQL
   IF g_properties.getValue("type_id") = "q01" OR
      g_properties.getValue("type_id") = "q02" OR
      g_properties.getValue("type_id") = "q03" OR
      g_properties.getValue("type_id") = "q04" THEN
      CALL g_properties.addAttribute("mdl_mark", "#該樣板不需此段落")
      CALL g_properties.addAttribute("mdl_mark2", "#該樣板不需此段落")
   END IF
   #200804-00006 --- end ---
   LET ls_tmp = adzp152_make_slice("a26")
   CALL g_properties.addAttribute("general_main_function", ls_tmp)

   #YSC-E30002 ---start---
   #因為部分樣板可能沒有設定dataset的資訊，而造成部分程式段組不出來，因此要分開處理
   #YSC-E30001 ---start---
   LET l_master_dataset = g_properties.getValue("master_dataset_"||g_properties.getValue("type_id")||"_exist")
   IF NOT cl_null(l_master_dataset) THEN
      CASE
          WHEN g_properties.getValue("type_id") = "q01" OR g_properties.getValue("type_id") = "q03"

          WHEN g_properties.getValue("type_id") = "q04"
             IF l_master_dataset = "Y" THEN
                LET ls_tmp = adzp152_make_slice("qs17")
                CALL g_properties.addAttribute("master_fetch_cursor_action", ls_tmp)

                #YSC-E80002 ---start---
                #有dataset的狀況下，才取的到table的pk欄位
                IF l_master_dataset = "Y" THEN
                   LET ls_tmp = adzp152_make_slice("qs26")
                   CALL g_properties.addAttribute("default_search_field",ls_tmp)
                END IF
                #YSC-E80002 --- end ---
             ELSE
                LET ls_tmp = adzp152_make_slice("qs18")
                CALL g_properties.addAttribute("master_fetch_cursor_action", ls_tmp)
             END IF
      END CASE
   END IF
   # YSC-E30001 --- end ---

   LET l_detail_dataset = g_properties.getValue("detail_dataset_"||g_properties.getValue("type_id")||"_exist")
   IF NOT cl_null(l_detail_dataset) THEN
      CASE
          WHEN g_properties.getValue("type_id") = "q01" OR g_properties.getValue("type_id") = "q03"
             IF l_detail_dataset = "Y" THEN
                LET ls_tmp = adzp152_make_slice("qs04")
                CALL g_properties.addAttribute("detail_b_fill_data", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs05")
                CALL g_properties.addAttribute("detail_b_fill_relate_data", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs06")
                CALL g_properties.addAttribute("detail_b_fill_cursor_free", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs07")
                CALL g_properties.addAttribute("detail_b_fill2_data", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs08")
                CALL g_properties.addAttribute("detail_filter_dialog", ls_tmp)

                #YSC-E80002 ---start---
                #有dataset的狀況下，才取的到table的pk欄位
                LET ls_tmp = adzp152_make_slice("qs27")
                CALL g_properties.addAttribute("default_search_field",ls_tmp)
                #YSC-E80002 --- end ---
             ELSE
                LET ls_tmp = adzp152_make_slice("qs09")
                CALL g_properties.addAttribute("detail_b_fill_data", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs10")
                CALL g_properties.addAttribute("detail_b_fill_relate_data", ls_tmp)
                LET ls_tmp = ""
                CALL g_properties.addAttribute("detail_b_fill_cursor_free", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs11")
                CALL g_properties.addAttribute("detail_b_fill2_data", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs12")
                CALL g_properties.addAttribute("detail_filter_dialog", ls_tmp)
             END IF

          #YSC-E80002 ---start---
          WHEN g_properties.getValue("type_id") = "q02"
             #有dataset的狀況下，才取的到table的pk欄位
             LET ls_tmp = adzp152_make_slice("qs27")
             CALL g_properties.addAttribute("default_search_field",ls_tmp)
          #YSC-E80002 --- end ---

         WHEN g_properties.getValue("type_id") = "q04"
             IF l_detail_dataset = "Y" THEN
                #q04的pk欄位是取單頭而非單身
             ELSE
                LET ls_tmp = adzp152_make_slice("qs09")
                CALL g_properties.addAttribute("detail_b_fill_data", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs11")
                CALL g_properties.addAttribute("detail_b_fill2_data", ls_tmp)
                LET ls_tmp = adzp152_make_slice("qs12")
                CALL g_properties.addAttribute("detail_filter_dialog", ls_tmp)
             END IF
      END CASE

      #YSC-E30001 ---start---
      IF g_properties.getValue("type_id") = "q01" THEN
         IF g_properties.getValue("type_id_t") = "q01" THEN   #YSC-E40001  #tab檔紀錄的樣板
            # YSC-E40001  只有q01樣板才有資料過濾(filter)功能
            LET ls_tmp = adzp152_make_slice("qs13")
            CALL g_properties.addAttribute("general_filter", ls_tmp)
            LET ls_tmp = adzp152_make_slice("qs14")
            CALL g_properties.addAttribute("general_filter_parser", ls_tmp)
            LET ls_tmp = adzp152_make_slice("qs15")
            CALL g_properties.addAttribute("general_filter_show", ls_tmp)
            LET ls_tmp = adzp152_make_slice("qs16")
            CALL g_properties.addAttribute("general_filter_on_action", ls_tmp)

         ELSE
            # YSC-E40001  只有q03樣板才有tree節點開關的功能
            LET ls_tmp = adzp152_make_slice("qs20")
            CALL g_properties.addAttribute("general_tree_on_action", ls_tmp)
            LET ls_tmp = adzp152_make_slice("qs21")
            CALL g_properties.addAttribute("general_tree_expand", ls_tmp)
            LET ls_tmp = adzp152_make_slice("qs22")
            CALL g_properties.addAttribute("general_tree_desc_show", ls_tmp)
            LET ls_tmp = adzp152_make_slice("qs23")
            CALL g_properties.addAttribute("general_tree_chk_isnode", ls_tmp)
         END IF
      END IF
      #YSC-E30001 --- end ---

      # YSC-E70001 ---start---
      # 畫面若有設定sel欄位，需賦予相關action功能
      IF g_properties.getValue("detail_sel_set") = "Y" THEN
         LET ls_tmp = adzp152_make_slice("qs19")
         CALL g_properties.addAttribute("general_sel_action", ls_tmp)
      END IF
      # YSC-E70001 --- end ---
   END IF
   #YSC-E30002 --- end ---

END FUNCTION

#+ page轉換為page_id
PUBLIC FUNCTION adzp152_page_trans(ps_page)
   DEFINE ps_page       STRING
   DEFINE ls_value      STRING
   DEFINE ls_name       STRING
      
   LET ls_name = adzp152_create_name(ps_page, "detail_page_id", "<<<")
   LET ls_value = g_properties.getValue(ls_name)

   RETURN ls_value
   
END FUNCTION

# YSC-E30001 ---start---
#+ 左側qbe所需資訊(q04)
PUBLIC FUNCTION adzp152_make_qbe()
   DEFINE ls_fields       STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_bfield       STRING

   #master_vars_form(對應到單頭的畫面欄位名稱)
#  LET ls_fields = g_properties.getValue("master_all_field")
   LET ls_fields = g_properties.getValue("master_fields_all")

   #加上b_(此處須使用對應表取得)
   LET lst_token = base.StringTokenizer.create(ls_fields, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      #LET ls_bfield = ls_bfield,g_properties.getValue(ls_token), ','
      LET ls_bfield = ls_bfield, ls_token, ',' #因無對應表, 暫取原本欄位
   END WHILE

   LET ls_bfield = ls_bfield.subString(1,ls_bfield.getLength()-1)

   CALL g_properties.addAttribute("master_vars_form", ls_bfield)

END FUNCTION
# YSC-E30001 --- end ---

# YSC-E30002 ---start---
#+ 建立temptable
PRIVATE FUNCTION adzp152_create_temp_table()

   #Create temp table 畫面欄位資訊暫存檔
   CREATE TEMP TABLE adzp152_form_tmp
      (
         type         VARCHAR(10),  # 類型 (head / body)
         page         SMALLINT,     # 頁籤
         sr_record    VARCHAR(20),  # 螢幕陣列名稱
         tot_count    SMALLINT,     # 該螢幕陣列的總欄位數
         var_field    VARCHAR(50),  # 欄位代號
         form_field   VARCHAR(50)   # 畫面欄位代號
      );

   # YSC-E70003 ---start---
   #Create temp table 串查功能傳入參數資訊暫存檔
   CREATE TEMP TABLE adzp152_para_tmp
      (
         field          VARCHAR(100), # 欄位
         para_no        SMALLINT,     # 參數項次
         para_data      VARCHAR(200), # 參數值
         para_totcount  SMALLINT      # 參數個數
      );
   # YSC-E70003 --- end ---

END FUNCTION

PRIVATE FUNCTION adzp152_form_field_search(ls_n)
   DEFINE ls_n                om.DomNode
   DEFINE lst_token           base.StringTokenizer
   DEFINE ls_token            STRING
   DEFINE ls_type             LIKE type_t.chr10
   DEFINE ls_page             LIKE type_t.num5
   DEFINE ls_sr_record        LIKE type_t.chr20
   DEFINE ls_field_count      LIKE type_t.num5
   DEFINE ls_var_field        LIKE type_t.chr50
   DEFINE ls_form_field       LIKE type_t.chr50
   DEFINE ls_field_count_str  STRING
   DEFINE ls_str              STRING
   DEFINE ls_field_list       STRING
   DEFINE ls_error            LIKE type_t.num5
   DEFINE lit_token           base.StringTokenizer
   DEFINE li_token            STRING
   DEFINE ls_page_field       STRING
   DEFINE ls_qbe_field        STRING
   DEFINE ls_form_field_str   STRING
   DEFINE ls_form_field_list  STRING

   LET ls_error = FALSE

   IF ls_n.getAttribute("id") = "form_field" THEN
      LET ls_n = ls_n.getFirstChild()
      WHILE ls_n IS NOT NULL
         LET ls_type = NULL
         LET ls_sr_record = NULL
         LET ls_page = 0
         LET ls_field_count = 0
         LET ls_form_field_list = NULL

         LET ls_type = ls_n.getAttribute("id")
         IF ls_type = "head" THEN
            LET ls_field_count_str = "browser_field_count"
            LET ls_form_field_str = "master_form_field_list"
         ELSE
            LET ls_field_count_str = "browser_field_count",ls_page USING "<<<"
            LET ls_form_field_str = "detail_form_field_list",ls_page USING "<<<"

            LET ls_page = ls_n.getAttribute("page")
            LET ls_sr_record = ls_n.getAttribute("record")
            LET ls_str = "browser_screen_record",ls_page USING "<<<"
            CALL g_properties.addAttribute(ls_str, ls_sr_record)
         END IF
 
         LET ls_field_list = ls_n.getAttribute("value")

         # 擷取畫面欄位資訊
         LET lst_token = base.StringTokenizer.create(ls_field_list, ',')
         LET ls_field_count = lst_token.countTokens()
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            IF ls_token.getIndexOf("(",1) > 0 THEN
               LET ls_var_field = NULL
               LET ls_form_field = NULL

               # 欄位代號
               LET ls_var_field = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
               LET ls_var_field = adzp152_field_filter(ls_var_field,"field")

               # 畫面欄位代號
               LET ls_form_field = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf(")",1)-1)
               IF NOT cl_null(ls_form_field_list) THEN
                  LET ls_form_field_list = ls_form_field_list,","
               END IF
               LET ls_form_field_list = ls_form_field_list , ls_form_field

               # 寫入畫面欄位對應暫存檔中
               INSERT INTO adzp152_form_tmp
                  (type, page, sr_record, tot_count, var_field, form_field)
                  VALUES (ls_type, ls_page, ls_sr_record, ls_field_count, ls_var_field, ls_form_field)
               IF SQLCA.SQLCODE THEN
                  LET ls_error = TRUE
                  DISPLAY "ERROR: INSERT TEMPTABLE(adzp152_form_tmp)",SQLCA.SQLCODE
                  EXIT WHILE
               END IF
            END IF
         END WHILE

         IF ls_error THEN
            EXIT WHILE
         END IF
         # 紀錄每個table的欄位數量
         CALL g_properties.addAttribute(ls_field_count_str, ls_field_count)

         # 紀錄畫面欄位
         CALL g_properties.addAttribute(ls_form_field_str, ls_form_field_list)

         LET ls_n = ls_n.getNext()  #同一層 next sibling
      END WHILE
   END IF
END FUNCTION
# YSC-E30002 --- end ---

#外部參數排序
FUNCTION adzp152_arg_idx(pi_fix)
   DEFINE pi_fix        INTEGER
   DEFINE li_idx        INTEGER
   DEFINE li_idx2       INTEGER
   DEFINE li_cnt        INTEGER
   DEFINE ls_name       STRING
   DEFINE ls_value      STRING

   #排除固定參數外, 重新確定外部參數序號
   FOR li_idx = 1 TO 100
      LET li_idx2 = li_idx + pi_fix
      LET ls_name  = "general_para", li_idx USING "&&"
      LET ls_value = li_idx2 USING "&&"
      CALL g_properties.addAttribute(ls_name,ls_value)
   END FOR

END FUNCTION


#寫入樣板相關資訊
PUBLIC FUNCTION adzp152_update_gzzx(ps_line)
   DEFINE ps_line  STRING      #資訊字串(未解析)
   DEFINE lr_gzzx  RECORD LIKE gzzx_t.*
   DEFINE ls_ver   STRING
   DEFINE li_b     INTEGER
   DEFINE li_e     INTEGER
   DEFINE li_cnt   INTEGER

   #取得樣板代碼
   LET lr_gzzx.gzzx011 = g_properties.getValue("type_id")

   #取得樣板版本
   LET li_b = ps_line.getIndexOf('(Version:',1) + 8
   LET li_e = ps_line.getIndexOf(')',li_b)
   LET ls_ver = ps_line.subString(li_b+1,li_e-1)
   LET ls_ver = ls_ver.trim()
   LET lr_gzzx.gzzx012 = ls_ver

   #指定程式名稱
   LET lr_gzzx.gzzx001 = g_properties.getValue("app_id")

   #先檢查程式有註冊否
   LET li_cnt = 0
   SELECT COUNT(*) INTO li_cnt FROM gzzx_t
    WHERE gzzx001 = lr_gzzx.gzzx001

   IF li_cnt = 0 THEN
      DISPLAY 'ERROR(0): 該程式未註冊，請先於azzi900進行註冊！'
          EXIT PROGRAM
   ELSE
      UPDATE gzzx_t
         SET gzzx011 = lr_gzzx.gzzx011,
             gzzx012 = lr_gzzx.gzzx012
       WHERE gzzx001 = lr_gzzx.gzzx001
          IF SQLCA.sqlcode THEN
             DISPLAY 'WARNING: 樣板資訊回寫gzzx_t失敗!'
          END IF
   END IF

END FUNCTION

#產生相關的遮罩段落
FUNCTION adzp152_create_all_mask()
   DEFINE ls_tmp   STRING
   DEFINE ls_title STRING

   LET ls_title = 'general_mask_funcs'
   LET ls_tmp   = '&include "erp/',g_properties.getValue("module")
                              ,'/',g_properties.getValue("app_id"),'_mask.4gl"'
   CALL g_properties.addAttribute(ls_title,ls_tmp)

END FUNCTION

#190108-00038 ---start---
#處理簡易畫面相關段落
FUNCTION adzp152_simple_mode()
   DEFINE ls_gzza001     LIKE gzza_t.gzza001   #程式代號
   DEFINE ls_dzzc004     LIKE dzzc_t.dzzc004   #簡易畫面實際存放的模組
   DEFINE ls_cnt         LIKE type_t.num5
   DEFINE ls_name        STRING
   DEFINE ls_value       STRING
   DEFINE ls_code        STRING
   DEFINE ls_module      STRING


   #確定是否有設定簡易畫面，先預設是未設定簡易畫面的模式
   LET g_simple_type = 'N'
   CALL g_properties.addAttribute("general_simple_mode_module","") #simple mode module

   LET ls_gzza001 = g_properties.getValue("app_id")   #程式代號
   LET ls_cnt = 0
   SELECT COUNT(1) INTO ls_cnt FROM dzzc_t
    WHERE dzzcstus = 'Y'
      AND dzzc001 = ls_gzza001

   #確定有設定簡易畫面的程式才需要產出以下相關的段落
   IF ls_cnt > 0 THEN
      #取出簡易畫面實際存放的模組
      #註：因為簡易畫面是透過設定而來，所以有可能標準化面已客製，但簡易畫面仍為標準的狀況
      #    所以簡易畫面需以dzzc_t中紀錄的實際模組(dzzc004)為主
      LET ls_dzzc004 = NULL
      SELECT dzzc004 INTO ls_dzzc004 FROM dzzc_t
       WHERE dzzcstus = 'Y'
         AND dzzc001 = ls_gzza001

      IF NOT cl_null(ls_dzzc004) THEN
         LET g_simple_type = 'Y'
         LET ls_dzzc004 = DOWNSHIFT(ls_dzzc004)
         CALL g_properties.addAttribute("general_simple_mode_module",ls_dzzc004) #simple mode module

         #產生簡易畫面切換的ON ACTION段落
         LET ls_name  = "general_on_action_ui_change"
         LET ls_value = adzp152_create_action_ui_change('dialog','a72')
         CALL g_properties.addAttribute(ls_name,ls_value)

         IF g_properties.getValue("type_id") = "i01" OR
            g_properties.getValue("type_id") = "i05" THEN
            LET ls_name  = "general_on_action_ui_change_menu"
            LET ls_value = adzp152_create_action_ui_change('menu','a72')
            CALL g_properties.addAttribute(ls_name,ls_value)
         END IF

         #產生確認畫面模式的FUNCTION段
         LET ls_name = "general_ui_chk"
         LET ls_value = adzp152_make_slice('a73')
         CALL g_properties.addAttribute(ls_name,ls_value)

         #產生init()段更換圖片及說明的程式段
         LET ls_name = "general_define_ui_change_icon"
         LET ls_value = adzp152_make_slice('a74')
         CALL g_properties.addAttribute(ls_name,ls_value)

         #產生轉換畫面時相關程式段
         LET ls_name = "general_ui_change"
         LET ls_value = adzp152_make_slice('a75')
         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF
   END IF

   LET ls_code = g_properties.getValue("general_prefix")
   LET ls_module = g_properties.getValue("general_module")

   #產生main段開啟畫面的相關段落
   LET ls_name = 'general_main_open_window'
   IF g_simple_type = 'Y' THEN
      #因為1.0與3.0是共用同一份程式碼，
      #因此有可能在3.0環境下有設了3.0版才有的功能，
      #但是因為1.0與3.0是讀同一份設定檔 (因為資料庫是同一份，沒有區分版本)
      #所以切換到1.0環境後，若單純只讀設定檔的話，會判斷是有設定的，
      #為避免在1.0環境下執行到3.0才有的功能，
      #必須在產出相關的程式段時，一併產出判斷版本的部分
      LET ls_value =          "&ifdef v1   #區分版本\n",
                     6 SPACES,"   OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_module,"',g_code)\n",
                     6 SPACES,"&else\n",
                     6 SPACES,"   IF g_simple_mode THEN   #簡化模式\n",
                     6 SPACES,"      OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_dzzc004,"',g_simple_name)\n",
                     6 SPACES,"   ELSE\n",
                     6 SPACES,"      OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_module,"',g_code)\n",
                     6 SPACES,"   END IF\n",
                     6 SPACES,"&endif"
   ELSE
      LET ls_value = "OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_module,"',g_code)"
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)

   #產生main段確認畫面模式的段落
   IF g_simple_type = 'Y' THEN
      LET ls_name = "general_check_ui_mode"
      LET ls_value =          "&ifndef v1   #區分版本\n",
                     3 SPACES,"   #確認畫面模式\n",
                     3 SPACES,"   CALL ",ls_code,"_ui_chk()\n",
                     3 SPACES,"&endif"
      CALL g_properties.addAttribute(ls_name,ls_value)
   END IF

END FUNCTION
#190108-00038 --- end ---
