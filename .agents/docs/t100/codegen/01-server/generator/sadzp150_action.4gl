#adzp150 副程式 - action相關產生
# code_version.: 5.25.02-2012.04.16(0000-0001)
# Memo : 180411-00062#1 by yscheng 修正若有page刪除，page id隨著變化，會導致add-point錯誤的問題
#        180611-00048#1 by yscheng ITM類程式產生器(adzp150)偵測到作業代碼第四碼為Q者，須加上datarefresh的功能
#        181128-00048#1 by 00824   3.0功能回追1.0環境(多語言開窗改為不區分版本)
#        190108-00038#1 by 00824   簡易畫面功能
#        190703-00069#1 by 02667   azzi880產生問題修復
#        210810-00005#1 by 00824   datarefresh處需讓指標(g_detail_idx, g_detail_idx2)歸正為1
#        210927-00001#1 by 00824   i02樣板無定義g_detail_idx2變數，修正210810-00005錯誤
#        211022-00003#1 by 05598   a43樣板調整模組路徑產生段邏輯，以gzza003欄位為模組路徑
#        230710-00021#1 by 00824   修正若是output為第一個acton的話，add-point名稱會有缺漏的問題

SCHEMA ds
GLOBALS "../4gl/adzp150_global.inc"   #180108-00012

GLOBALS
   DEFINE g_properties   om.SaxAttributes
   DEFINE g_component_list DYNAMIC ARRAY OF RECORD
          comp_id    STRING,
          action     LIKE type_t.num5,
          standard   LIKE type_t.num5,
          including  LIKE type_t.num5
          END RECORD
   DEFINE ga_action DYNAMIC ARRAY OF STRING
   DEFINE g_update       STRING
END GLOBALS

DEFINE g_dzbb RECORD 
         prog_name    STRING,
         point_name   STRING,
         point_ver    STRING,
         addpoint     STRING
         END RECORD


#+ 產生 ON ACTION
PUBLIC FUNCTION adzp150_write_on_action(li_pos,p_type,ldnode_parent,p_loc,p_page)
   DEFINE p_loc           STRING                #所在位置
   DEFINE p_type          STRING                #樣板型態
   DEFINE p_page          STRING                #頁數
   DEFINE ldnode_parent   om.DomNode
   DEFINE ldnode_action   om.DomNode
   DEFINE li_cnt          LIKE type_t.num10
   DEFINE li_pos          LIKE type_t.num10     #空幾格
   DEFINE ls_m_action     STRING
   DEFINE ls_m_when       STRING
   DEFINE ls_single       STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_m_action_bak STRING                #舊值備份 預防重複出現
   DEFINE ls_m_when_bak   STRING                #舊值備份 預防重複出現
   DEFINE ls_single_bak   STRING                #舊值備份 預防重複出現
   DEFINE ls_page         STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING


   #取得頁數代碼
   LET ls_page = p_page.trim()
   
   FOR li_cnt = 1 TO ldnode_parent.getChildCount()
      LET ldnode_action = ldnode_parent.getChildByIndex(li_cnt)
      
      LET ga_action[li_cnt] = ldnode_action.getAttribute("id")
      #LET ls_tmp = adzp150_on_action(li_pos+1,ldnode_action,p_loc,ldnode_parent.getAttribute("page"))
      #180411-00062 ---modify start---
      #之前有針對page被刪除，導致若直接取current page時，會造成add-point名稱錯亂的問題，
      #因此須改為取page_id，因為page_id不會因為page調整而改變，
      #但之前調整時漏掉了這個部分，所以現在補上，
      #但是1.0版本若要調整會造成更大問題，因此決議只改在3.0版本
      #181128-00048 ---mark start---
      #多語言開窗改為不區分版本，但因為上述原因，因此只能維持取current page的方式，而非取page_id
   #  IF g_erp_ver = "1.0" THEN
         LET ls_tmp = adzp150_make_action(ldnode_action,p_loc,ldnode_parent.getAttribute("page"))
   #  ELSE
   #     LET ls_tmp = adzp150_make_action(ldnode_action,p_loc,adzp150_page_trans(ldnode_parent.getAttribute("page")))
   #  END IF
      #181128-00048 --- mark end ---
      #180411-00062 --- modify end ---
      IF (ldnode_action.getAttribute("id") = 'insert' OR
         ldnode_action.getAttribute("id") = 'delete'  OR
         ldnode_action.getAttribute("id") = 'reproduce') AND
         g_properties.getValue("type_id") = "i05" THEN
         LET ls_tmp = cl_replace_str(ls_tmp,"_insert()","_insert(DIALOG)")
         LET ls_tmp = cl_replace_str(ls_tmp,"_reproduce()","_reproduce(DIALOG)")
         LET ls_tmp = cl_replace_str(ls_tmp,"_delete()","_delete(DIALOG)")
      END IF 
      LET ls_single = ls_single, "\n",ls_tmp
      IF ldnode_action.getAttribute("id") = "modify" AND
         ( 
         g_properties.getValue("type_id") = "i02" OR
         g_properties.getValue("type_id") = "i04" OR
         g_properties.getValue("type_id") = "i07" OR
         g_properties.getValue("type_id") = "i08" OR
         g_properties.getValue("type_id") = "t01" OR
         g_properties.getValue("type_id") = "t02" 
         )
         THEN
         LET ls_tmp = cl_replace_str(ls_tmp,'modify','modify_detail') 
         LET ls_tmp = cl_replace_str(ls_tmp,'_modify_detail()','_modify()') 
         LET ls_tmp = cl_replace_str(ls_tmp,'cl_auth_chk_act("modify_detail")','cl_auth_chk_act("modify")') 
         LET ls_tmp = cl_replace_str(ls_tmp,"LET g_aw = ''","LET g_aw = g_curr_diag.getCurrentItem()")
         LET ls_single = ls_single, "\n", ls_tmp
      END IF    
      
   END FOR

   CASE p_loc

      WHEN "menu"
         CALL g_properties.addAttribute("general_action_call",ls_single)
         LET ls_tmp = ls_single
         LET ls_tmp = cl_replace_str(ls_tmp,"EXIT DIALOG","EXIT MENU")
         LET ls_tmp = cl_replace_str(ls_tmp,"menu.","menu2.")
         IF g_properties.getValue("type_id") = "i05" AND
            g_properties.getValue("app_id") = "azzi880" #暫時用
            THEN
            #190703-00069#1 -S-
            #LET ls_tmp = cl_replace_str(ls_tmp,"_insert(DIALOG)","_insert('')")
            #LET ls_tmp = cl_replace_str(ls_tmp,"_reproduce(DIALOG)","_reproduce('')")
            #LET ls_tmp = cl_replace_str(ls_tmp,"_delete(DIALOG)","_delete('')")
            LET ls_tmp = cl_replace_str(ls_tmp,"_insert(DIALOG)","_insert(NULL)")
            LET ls_tmp = cl_replace_str(ls_tmp,"_reproduce(DIALOG)","_reproduce(NULL)")
            LET ls_tmp = cl_replace_str(ls_tmp,"_delete(DIALOG)","_delete(NULL)")  
            #190703-00069#1 -E_
         END IF 
         CALL g_properties.addAttribute("general_action_call_in_menu",ls_tmp)

      WHEN "detail_show"
         LET ls_name = "detail_saction_choice",ls_page
         LET ls_tmp = g_properties.getValue(ls_name)
         LET ls_tmp = ls_tmp, "\n", ls_single
         LET ls_value = ls_tmp
         CALL g_properties.addAttribute(ls_name,ls_value)

      WHEN "master_input"
         CALL g_properties.addAttribute("master_iaction_call",ls_single)

      WHEN "detail_input"
         LET ls_name = adzp150_create_name(ls_page, "detail_iaction_call", "<<<")
         LET ls_tmp = g_properties.getValue(ls_name)
         LET ls_tmp = ls_tmp, "\n", ls_single
         LET ls_value = ls_tmp
         CALL g_properties.addAttribute(ls_name,ls_value)
         
   END CASE

END FUNCTION

#+ 產生action段落(讀取slice)
PUBLIC FUNCTION adzp150_make_action(pdnode_action,p_loc,ps_page)
   DEFINE pdnode_action   om.DomNode #action tag
   DEFINE ldnode_sub      om.DomNode #action tag(sub)
   DEFINE p_loc           STRING     #該action所在位置
   DEFINE ps_page         STRING     #該action所在page
   DEFINE ls_action_name  STRING     #action名稱
   DEFINE ls_tmp          STRING
   DEFINE ls_tmp2         STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   DEFINE ls_mode         STRING     #action類型
   DEFINE ls_prog         STRING     #action串查(程式名稱)
   DEFINE ls_para         STRING     #action串查(參數)
   DEFINE ls_chk          STRING     #action是否進行身分檢驗
   DEFINE ls_lib          STRING     #action串查(lib名稱)
   DEFINE ls_prog_type    STRING     #程式樣板
   DEFINE ls_return       STRING
   DEFINE ls_adp_name     STRING
   DEFINE ls_pre          STRING
   DEFINE ls_gzza003      LIKE gzza_t.gzza003 #211022-00003 add
   DEFINE ls_gzza001      LIKE gzza_t.gzza001 #211022-00003 add
   DEFINE li_idx          INTEGER    #161109-00070 #1 add
   DEFINE ls_tbl_pre      STRING     #161109-00070 #1 add
   
   #先抓action名稱&mode類型
   LET ls_action_name = pdnode_action.getAttribute("id")
   LET ls_mode        = pdnode_action.getAttribute("mode")
   LET ls_prog        = pdnode_action.getAttribute("prog")
   LET ls_lib         = pdnode_action.getAttribute("lib")
   LET ls_para        = pdnode_action.getAttribute("parameter")
   LET ls_chk         = pdnode_action.getAttribute("chk")
   LET ls_prog_type   = g_properties.getValue("type_id")
   
   #判斷是否為標準action 並填入 mdl_action_call
   LET ls_name  = "mdl_action_call"
   LET ls_value = ""
   IF ls_mode = "popup" THEN
      LET ldnode_sub = pdnode_action.getFirstChild()
      WHILE ldnode_sub IS NOT NULL
         LET ls_value = ls_value, adzp150_make_action(ldnode_sub,p_loc,ps_page||"_sub")
         LET ldnode_sub = ldnode_sub.getNext()
      END WHILE
      CALL g_properties.addAttribute(ls_name,ls_value)
	  #mdl_menu_name
	  CALL g_properties.addAttribute('mdl_menu_name',ls_action_name)
      LET ls_value = adzp150_make_slice("a44")
      LET ls_value = cl_replace_str(ls_value,"         ","               ")
   ELSE
      IF adzp150_standard_actions(ls_action_name) THEN
         LET ls_value = "CALL ",g_properties.getValue("general_prefix"),"_",ls_action_name,"()"
      ELSE
         LET ls_value = adzp150_parse_element(pdnode_action,1)
      END IF
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #進行資料填入, mdl_action_name
   LET ls_name  = "mdl_action_name"
   LET ls_value = ls_action_name
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #230710-00021 ---start---  
   #判斷所在位置並(mdl_action_loc)決定離開時的名稱(mdl_exit)
   LET ls_name = "mdl_action_loc"
   CASE p_loc
      WHEN "menu"
         LET ls_value = "menu"
      WHEN "master_input"
         LET ls_value = "input.master_input"
      WHEN "detail_show"
         LET ls_value = "menu.detail_show.page",ps_page
      WHEN "detail_input"
         LET ls_value = "input.detail_input.page",ps_page
   END CASE
   CALL g_properties.addAttribute(ls_name,ls_value)
   #230710-00021 --- end ---

   #output特別處理
   LET ls_name  = "mdl_output"
   LET ls_tmp  = g_properties.getValue("app_id")
   LET ls_tmp2 = g_properties.getValue("type_id_t")
   IF (ls_action_name = "output" OR ls_action_name = "quickprint") AND      #output專有   #161130-00002 新增quickprint
      ls_tmp.subString(4,4) = 't' AND    #第四碼為t
      ls_tmp2.getIndexOf('c',1) = 0 AND  #不可為子作業
      ls_tmp2 <> 'p00' THEN              #不可為p00
      LET ls_pre = g_properties.getValue("app_id")
      #LET ls_pre = ls_pre.subString(1,3)                                   # 211022-00003 mark
      # 211022-00003 add s
      LET ls_gzza001 = ls_pre
      SELECT gzza003 INTO ls_gzza003 FROM gzza_t
       WHERE gzza001 = ls_gzza001 AND gzzastus ='Y'
         AND gzza002 ='T'
      LET ls_pre = DOWNSHIFT(ls_gzza003)
      # 211022-00003 add e
      LET ls_value = '&include "erp/',ls_pre,'/',g_properties.getValue("app_id"),'_rep.4gl"\n',
                     '               #add-point:ON ACTION ',g_properties.getValue("mdl_action_name"),'.after name="',g_properties.getValue("mdl_action_loc"),'.after_',g_properties.getValue("mdl_action_name"),'"\n',
                     '               {<point name="',g_properties.getValue("mdl_action_loc"),'.after_',g_properties.getValue("mdl_action_name"),'" />}\n',
                     '               #END add-point'
					 
   ELSE
      LET ls_value = "#該樣板不需此段落"
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #mdl_modify_detail (只有modify需要)
   LET ls_name  = "mdl_modify_detail"
   IF ls_action_name = "modify" THEN
      LET ls_value = "LET g_aw = ''"
   ELSE
      LET ls_value = "#該樣板不需此段落"
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #230710-00021 ---mark start---
#  #判斷所在位置並(mdl_action_loc)決定離開時的名稱(mdl_exit)
#  LET ls_name = "mdl_action_loc"
#  CASE p_loc
#     WHEN "menu"
#        LET ls_value = "menu"
#     WHEN "master_input"
#        LET ls_value = "input.master_input"
#     WHEN "detail_show"
#        LET ls_value = "menu.detail_show.page",ps_page
#     WHEN "detail_input"
#        LET ls_value = "input.detail_input.page",ps_page
#  END CASE
#  CALL g_properties.addAttribute(ls_name,ls_value)
   #230710-00021 --- mark end ---

   #mdl_exit
   LET ls_name = "mdl_exit"
   LET ls_value = "" 
   CASE p_loc
      WHEN "menu"
         IF ls_action_name = "query" THEN
            IF g_properties.getValue("type_id") = "i02" OR 
               g_properties.getValue("type_id") = "i04" OR 
               g_properties.getValue("type_id") = "i07" OR 
               g_properties.getValue("type_id") = "i08" OR 
               g_properties.getValue("type_id") = "t01" OR 
               g_properties.getValue("type_id") = "t02" THEN
               LET ls_value = adzp150_make_slice("a59")
            END IF
         END IF

      WHEN "master_input"
         LET ls_value = "#該樣板不需此段落"
      WHEN "detail_show"
         #LET ls_value = "EXIT DIALOG"
      WHEN "detail_input"
         LET ls_value = "#該樣板不需此段落"
      OTHERWISE
         LET ls_value = "#該樣板不需此段落"
   END CASE
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #popup特殊處理
   LET ls_name = "mdl_action_mark"
   IF ls_mode = "popup" THEN
      LET ls_value = "#該樣板不需此段落"
   ELSE
      LET ls_value = ""
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #確認是否要進行身分驗證(cl_auth_chk_act)
   LET ls_name = "mdl_chk_mark"
   IF ls_chk = 'N' AND NOT cl_null(ls_chk) THEN
      LET ls_value = "#該樣板不需此段落"
   ELSE
      LET ls_value = ""
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)

   #210511-00011 ---mark start---
   #功能尚未完整，暫時mark
   #190806-00005 ---start---
   #雙擊單身欄位游標直接進入雙擊位置(排除單檔) 3.0獨有功能
#  &ifdef v3   #210511-00011
#  LET ls_name = "mdl_detail_double_click"
#  IF ls_action_name = "modify" OR ls_action_name = "modify_detail" THEN
#     IF g_properties.getValue("type_id") = "i02" OR 
#        g_properties.getValue("type_id") = "i04" OR 
#        g_properties.getValue("type_id") = "i07" OR 
#        g_properties.getValue("type_id") = "i08" OR 
#        g_properties.getValue("type_id") = "t01" OR 
#        g_properties.getValue("type_id") = "t02" THEN

#        LET ls_value = "LET g_curfld = DIALOG.getCurrentItem()"
#     ELSE
#        LET ls_value = "#該樣板不需此段落"
#     END IF
#  ELSE
#  #  LET ls_value = ""
#     LET ls_value = "#該樣板不需此段落"
#  END IF
#  CALL g_properties.addAttribute(ls_name,ls_value)
#  &endif   #210511-00011
   #190806-00005 --- end ---
   #210511-00011 --- mark end ---

   #161109-00070 #1 ---modify start---
#  IF g_properties.getValue("type_id") = "i02" THEN   #170316-00030
#  IF g_properties.getValue("type_id") = "i02" OR g_properties.getValue("type_id") = "i07" THEN   #170316-00030   #170804-00012 mark
   IF g_properties.getValue("type_id") = "i02" OR g_properties.getValue("type_id") = "i07"   #170316-00030   #170804-00012 modify
      OR g_properties.getValue("type_id") = "t02" THEN   #170804-00012 modify
      CASE
         WHEN ls_action_name = "modify"
            LET ls_return = adzp150_make_slice("a67")

         WHEN ls_action_name = "delete"
            LET ls_return = adzp150_make_slice("a67")

         OTHERWISE
            LET ls_return = adzp150_make_slice("a43")
      END CASE
   ELSE
      LET ls_return = adzp150_make_slice("a43")
   END IF
   #161109-00070 #1 --- modify end ---
   
   LET ls_adp_name =  g_properties.getValue("mdl_action_loc"),'.',g_properties.getValue("mdl_action_name")

   #處理程式串查段落(若有定義prog)
   IF NOT cl_null(ls_prog) THEN
      CALL adzp150_make_prog_link(ls_prog,ls_para,ls_adp_name)
      #如果有menu2需額外填寫(i01,i05專用)
      IF (ls_prog_type = "i01" OR ls_prog_type = "i05" ) AND p_loc = 'menu' THEN
         CALL adzp150_make_prog_link(ls_prog,ls_para,"menu2."||ls_action_name)
      END IF
   END IF   
        
   #處理lib串查段落(若有定義prog)
   IF NOT cl_null(ls_lib) THEN
      CALL adzp150_make_lib_link(ls_lib,ls_para,ls_adp_name)
      #如果有menu2需額外填寫(i01,i05專用)
      IF (ls_prog_type = "i01" OR ls_prog_type = "i05" ) AND p_loc = 'menu' THEN
         CALL adzp150_make_lib_link(ls_lib,ls_para,"menu2."||ls_action_name)
      END IF
   END IF   
        
   RETURN ls_return    

END FUNCTION


#+ 確認該標準的action是否是定義成在樣板檔裡面就有的  Yes:不用組 FUNCTION xxx()
PRIVATE FUNCTION adzp150_standard_actions(ps_act_id) 

   DEFINE ps_act_id   STRING

   IF ps_act_id = "insert"    OR ps_act_id = "query"     OR 
      ps_act_id = "reproduce" OR ps_act_id = "modify"    OR 
      ps_act_id = "delete"    OR ps_act_id = "filter"    THEN
      CASE g_properties.getValue("type_id")
         WHEN "i01"

         WHEN "i02"
            IF ps_act_id ="reproduce" THEN
               DISPLAY "WARNING:",g_properties.getValue("type_id"),"樣板中不含",ps_act_id,"功能"
               RETURN FALSE
            END IF
            
         WHEN "i04"   

         WHEN "i05"   

         WHEN "i07"     
         WHEN "i08"     

         WHEN "t01" 
         
         WHEN "t02"
            IF ps_act_id ="reproduce" OR 
               #ps_act_id ="insert"    OR
               ps_act_id ="delete"    THEN
               DISPLAY "WARNING:",g_properties.getValue("type_id"),"樣板中不含",ps_act_id,"功能"
               RETURN FALSE
            END IF

      END CASE
      RETURN TRUE
   ELSE
      RETURN FALSE
   END IF

END FUNCTION

#+ 組合多語言架構會用到的action
PUBLIC FUNCTION adzp150_update_item_action(p_loc,p_node) 
   DEFINE p_node          om.DomNode
   DEFINE p_loc           STRING
   DEFINE ls_type         STRING
   DEFINE ls_iaction_call STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_page         STRING
   DEFINE li_idx          INTEGER
   DEFINE li_cnt          INTEGER
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   
   LET ls_type = p_node.getAttribute("type")
   
   #多語言才進行處理
   IF ls_type <> "lang" THEN
      RETURN
   END IF
   
   #針對各個欄位處理
   LET lst_token = base.StringTokenizer.create(p_node.getAttribute("field"), ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
   
      #若為單頭
      IF p_loc = 'm' THEN
         #LET ls_iaction_call =  g_properties.getValue("master_iaction_call"),
         #                       "   ON ACTION update_item \n\n"
                                 
         #LET ls_iaction_call =  g_properties.getValue("master_iaction_call"),
         #                       "   ON ACTION update_item INFIELD ",ls_token,"\n\n"
         #CALL g_properties.addAttribute("master_iaction_call",ls_iaction_call)
         RETURN 
      ELSE
      #若為單身
         #取得單身page所在
         LET li_cnt = g_properties.getValue("page")
         FOR li_idx = 1 TO li_cnt
            LET ls_tmp = "detail_field_loc",li_idx USING "<<<"
            LET ls_tmp = g_properties.getValue(ls_tmp)
            IF ls_tmp.getIndexOf(ls_token,1) > 0 THEN
               LET ls_page = li_idx USING "<<<"
               EXIT FOR
            END IF
         END FOR
         LET ls_iaction_call =  g_properties.getValue("detail_iaction_call"||ls_page),
                                "   ON ACTION update_item \n\n"
             
         #LET ls_iaction_call =  g_properties.getValue("detail_iaction_call"||ls_page),
         #                       "   ON ACTION update_item INFIELD ",ls_token,"\n\n"
         CALL g_properties.addAttribute("detail_iaction_call"||ls_page,ls_iaction_call)
         RETURN    
      END IF
   END WHILE

END FUNCTION

#+ action default產出段落
PUBLIC FUNCTION adzp150_create_action_default()
   DEFINE ls_default_action STRING
   DEFINE li_idx            INTEGER
   DEFINE li_cnt            INTEGER
   DEFINE lb_chk            BOOLEAN
   DEFINE ls_value          STRING
   DEFINE ls_name           STRING
   DEFINE ls_allow          STRING
   
   LET li_cnt = 0
   LET lb_chk = FALSE
   FOR li_idx = 1 TO  ga_action.getLength()

      #那些action可以啟動default_action, 不行則略過
      LET ls_allow = "insert"
      IF ls_allow.getIndexOf(ga_action[li_idx],1) = 0 THEN
         CONTINUE FOR
      END IF
      
      LET lb_chk = TRUE
   
      LET li_cnt = li_cnt + 1
   
      #mdl_action
      LET ls_name  = adzp150_create_name(li_cnt, "mdl_action", "<<<")
      LET ls_value = ga_action[li_idx]
      CALL g_properties.addAttribute(ls_name,ls_value)
    
      #mdl_para
      LET ls_name  = adzp150_create_name(li_cnt, "mdl_para", "<<<")
      IF (ls_value = 'insert' OR ls_value = 'delete' OR ls_value  = 'reproduce') AND
         g_properties.getValue("type_id")  = "i05" THEN
         LET ls_value = "DIALOG"
      ELSE
         LET ls_value = ""
      END IF 
      CALL g_properties.addAttribute(ls_name,ls_value)
      
   END FOR
   
   LET ls_name  = "mdl_mdls"
   LET ls_value = li_cnt USING "<<<"
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   IF lb_chk THEN
      LET ls_default_action = adzp150_make_slice("a42")
   END IF

   LET ls_name  = "general_action_default"
   LET ls_value = ls_default_action
   CALL g_properties.addAttribute(ls_name,ls_value)
   
END FUNCTION

#+ 產生串查段落(程式串查)
PUBLIC FUNCTION adzp150_make_prog_link(ps_prog,ps_para,ps_adp_name)
   DEFINE ps_prog         STRING
   DEFINE ps_para         STRING
   DEFINE ps_adp_name     STRING
   DEFINE ls_adp          STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_page_name    STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   DEFINE li_idx          INTEGER
   
   #填入mdl_prog(串查程式名稱)
   LET ls_name  = "mdl_prog"
   LET ls_value = ps_prog
   CALL g_properties.addAttribute(ls_name,ls_value)

   #填入mdl_para(傳遞參數), 相關動作處理
   IF cl_null(ps_para) THEN
   LET ls_name  = "mdl_link_mark"
   LET ls_value = "#"
      CALL g_properties.addAttribute(ls_name,ls_value)
   END IF
    
   #拆解並轉換
   LET li_idx = 1
   LET lst_token = base.StringTokenizer.create(adzp150_parse_para(ps_para), ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_name = adzp150_create_name(li_idx, "mdl_link_para", "<<<")
      LET ls_value = ls_token
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      LET li_idx = li_idx + 1
   END WHILE
   
   LET ls_name = "mdl_mdls"
   LET ls_value = (li_idx - 1) USING "<<<"
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   LET ls_adp = adzp150_make_slice("a41")
    
   #寫入adp
   LET g_dzbb.prog_name  = g_properties.getValue("app_id")
   LET g_dzbb.point_name = ps_adp_name
   LET g_dzbb.point_ver  = g_properties.getValue("general_adp_ver")
   LET g_dzbb.addpoint   = ls_adp
   CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))
   
END FUNCTION 

#+ 產生串查段落(lib串查)
PUBLIC FUNCTION adzp150_make_lib_link(ps_lib,ps_para,ps_adp_name)
   DEFINE ps_lib          STRING
   DEFINE ps_para         STRING
   DEFINE ps_adp_name     STRING
   DEFINE ls_adp          STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   
   #填入mdl_lib(串查lib名稱)
   LET ls_name  = "mdl_lib"
   LET ls_value = ps_lib
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #填入mdl_para(串查參數)
   LET ls_name  = "mdl_para"
   LET ls_value = adzp150_parse_para(ps_para)
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   LET ls_adp = adzp150_make_slice("a45")

   #寫入adp
   LET g_dzbb.prog_name  = g_properties.getValue("app_id")
   LET g_dzbb.point_name = ps_adp_name
   LET g_dzbb.point_ver  = g_properties.getValue("general_adp_ver")
   LET g_dzbb.addpoint   = ls_adp
   CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))
   
END FUNCTION 

#+ 解析參數, 轉為實際欄位(變數)
PUBLIC FUNCTION adzp150_parse_para(ps_para)
   DEFINE ps_para         STRING
   DEFINE ls_para         STRING
   DEFINE ls_token        STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_master_field STRING
   DEFINE ls_page_name    STRING
   
   #先取得單頭所有欄位
   LET ls_master_field = g_properties.getValue("master_all_field")
   
   #拆解並轉換
   LET lst_token = base.StringTokenizer.create(ps_para, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_master_field.getIndexOf(ls_token,1) THEN
         #先確定是否在單頭
         LET ls_para = ls_para, g_properties.getValue("master_var_title"),".",ls_token
      ELSE
         #若不在單頭則尋找單身
         LET ls_page_name = adzp150_find_page(ls_token,'')
         IF cl_null(ls_page_name) THEN
            #非單身欄位, 視為一般變數
            LET ls_para = ls_para, ls_token
         ELSE
            #單身欄位, 定義變數名稱
            LET ls_para = ls_para, adzp150_find_page(ls_token,''),".",ls_token
         END IF
      END IF

      IF lst_token.hasMoreTokens() THEN
         LET ls_para = ls_para, ','
      END IF
      
   END WHILE

   RETURN ls_para
   
END FUNCTION


#+ 產出相關文件action內容(包含新增)
PUBLIC FUNCTION adzp150_related_document(ps_loc,ps_type,ps_slice)
   DEFINE ps_loc           STRING
   DEFINE ps_type          STRING  
   DEFINE ps_slice         STRING  
   DEFINE ls_return        STRING
   DEFINE ls_name          STRING
   DEFINE ls_value         STRING
   DEFINE ls_field_keys    STRING
   DEFINE ls_var_keys      STRING
   DEFINE ls_token         STRING
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token2        STRING
   DEFINE lst_token2       base.StringTokenizer
   DEFINE li_idx           INTEGER
   DEFINE ls_docdt         STRING
   DEFINE ls_var_all       STRING
   
   #判斷要用的是單頭Key還是單身Key
   IF ps_type = 'm' THEN
      LET ls_field_keys = g_properties.getValue("master_field_allkeys")
      LET ls_var_keys   = g_properties.getValue("master_var_allkeys")
   ELSE
      LET ls_field_keys = g_properties.getValue("detail_field_allkeys")
      LET ls_var_keys   = g_properties.getValue("detail_var_allkeys")
      IF g_properties.getValue("type_id") = "t02" THEN
         LET ls_var_keys = cl_replace_str(ls_var_keys,'l_ac','g_detail_idx')
      END IF
   END IF

   #拆解並轉換
   LET li_idx = 0

   LET lst_token  = base.StringTokenizer.create(ls_field_keys, ',')
   LET lst_token2 = base.StringTokenizer.create(ls_var_keys, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token  = lst_token.nextToken()
      LET ls_token2 = lst_token2.nextToken()
      
      LET li_idx = li_idx + 1
      
      #mdl_field_pk
      LET ls_name  = adzp150_create_name(li_idx, "mdl_field_pk", "<<<")
      LET ls_value = ls_token
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #mdl_var_pk
      LET ls_name  = adzp150_create_name(li_idx, "mdl_var_pk", "<<<")
      LET ls_value = ls_token2
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #目前無法接受超出5個key
      #IF li_idx >= 5 THEN
      #   EXIT WHILE
      #END IF
      
   END WHILE
   
   #mdl_mdls
   LET ls_name  = "mdl_mdls"
   LET ls_value = li_idx USING "<<<"
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #mdl_addpoint
   LET ls_name  = "mdl_addpoint"
   LET ls_value = ps_loc #dialog or menu
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #mdl_docdt
   LET ls_docdt = g_properties.getValue("master_var_title"),'.',  #EX: g_itta_m
                  g_properties.getValue("master_tbl_prefix"), #EX: itta
                  "docdt"
   LET ls_var_all = g_properties.getValue("master_vars_all")
   
   IF ls_var_all.getIndexOf(ls_docdt,1) THEN
      LET ls_value = ls_docdt
   ELSE
      LET ls_value = "''"
   END IF
   LET ls_name  = "mdl_docdt"
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   LET ls_return = adzp150_make_slice(ps_slice)
   
   #160621
   LET ls_return = adzp150_make_slice(ps_slice)
   
   RETURN ls_return

END FUNCTION


#180611-00048 ---start---
#+ 產出datarefresh的ON ACTION 段落(只針對q類程式但用了itm類樣板的程式處理)
PUBLIC FUNCTION adzp150_create_action_datarefresh()
   DEFINE ls_tmp      STRING
   DEFINE ls_type     STRING
   DEFINE ls_name     STRING
   DEFINE ls_value    STRING
   DEFINE ls_spaces   STRING


   #若是q類程式，但是用了ITM類的樣板的程式，需補上datarefresh的功能
   LET ls_tmp = g_properties.getValue("app_id")   #取出程式代號
   LET ls_type = g_properties.getValue("type_id")   #取出樣板代號(實際使用的樣板)

   LET ls_name = "general_datarefresh"
   LET ls_value = NULL
   #判斷是否為q類的程式
   IF ls_tmp.subString(4,4) = "q" THEN
      IF ls_type NOT MATCHES "q0[0-4]" THEN   #判斷是否未使用Q類的樣板(Q01 - Q04)
         CASE
            #有單頭
            WHEN ls_type = "i01" OR ls_type = "i04" OR ls_type = "i05"
                 OR ls_type = "i07" OR ls_type = "i08" OR ls_type = "t01"

                 LET ls_spaces = 12 SPACES
                 LET ls_value = "ON ACTION datarefresh\n",
                                ls_spaces,"   LET g_action_choice = 'fetch'\n",
                                ls_spaces,"   CALL ${general_prefix}_fetch('F')"

            #無單頭
            WHEN ls_type = "i02" OR ls_type = "t02"
                 LET ls_spaces = 9 SPACES
                 LET ls_value = "ON ACTION datarefresh\n",
                                ls_spaces,"   LET g_detail_idx = 1\n"    #210810-00005  #210927-00001 modify
                 IF ls_type = "t02" THEN                                 #210927-00001
                    LET ls_value = ls_value,                             #210927-00001
                                ls_spaces,"   LET g_detail_idx2 = 1\n"   #210810-00005  #210927-00001 modify
                 END IF                                                  #210927-00001
                 LET ls_value = ls_value,                                #210927-00001
                                ls_spaces,"   CALL ${general_prefix}_b_fill(g_wc2)"
         END CASE

         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF
   END IF
END FUNCTION
#180611-00048 --- end ---

#190108-00038 ---start---
#+ 產出簡易畫面切換的ON ACTION 段落(3.0以上版本才有此功能)
PUBLIC FUNCTION adzp150_create_action_ui_change(ps_loc,ps_slice)
   DEFINE ps_loc           STRING
   DEFINE ps_slice         STRING
   DEFINE ls_return        STRING
   DEFINE ls_name          STRING
   DEFINE ls_value         STRING


   LET ls_return = NULL

   #mdl_addpoint
   LET ls_name  = "mdl_addpoint"
   LET ls_value = ps_loc   #dialog or menu
   CALL g_properties.addAttribute(ls_name,ls_value)

   LET ls_return = adzp150_make_slice(ps_slice)
   IF ps_loc = "menu" THEN
      LET ls_return = cl_replace_str(ls_return,'EXIT DIALOG','EXIT MENU')
   END IF

   RETURN ls_return
END FUNCTION
#190108-00038 --- end ---
