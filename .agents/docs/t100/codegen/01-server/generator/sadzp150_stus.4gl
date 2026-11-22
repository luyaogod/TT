#adzp150 副程式 - 狀態碼切換相關處理

IMPORT os
SCHEMA ds

GLOBALS "../4gl/adzp150_global.inc"   #180108-00012
GLOBALS
   DEFINE g_properties   om.SaxAttributes
   CONSTANT li_space = 3
   DEFINE g_update       STRING
END GLOBALS 

DEFINE g_dzbb RECORD 
         prog_name    STRING,
         point_name   STRING,
         point_ver    STRING,
         addpoint     STRING
         END RECORD

#+ create 狀態碼相關段落
PUBLIC FUNCTION adzp150_create_stus()
   DEFINE ls_stus STRING
   
   #確定有stus欄位
   LET ls_stus = g_properties.getValue("detail_stus_isExist")
   IF cl_null(ls_stus) OR ls_stus = "N" THEN
      CALL g_properties.addAttribute("general_mark_stus2","#該樣板不需此段落") 
   END IF 
   
   #確定有stus欄位
   LET ls_stus = g_properties.getValue("master_stus_isExist")
   IF cl_null(ls_stus) OR ls_stus = "N" THEN
      CALL g_properties.addAttribute("general_mark_stus","#該樣板不需此段落") 
      RETURN
   END IF 
   
   #準備資料
   CALL adzp150_read_stus()
   
   #簽核程式額外處理(a32~a37)
   CALL adzp150_create_signature()

   #a09,a10子樣板讀取&產生
   CALL adzp150_state_change()
   
   #a20子樣板讀取&產生
   CALL adzp150_state_control()
   
   #a21子樣板讀取&產生
   CALL adzp150_state_show()
   
   #a22子樣板讀取&產生
   CALL adzp150_state_action_control()
   
   #a24子樣板讀取&產生
   CALL adzp150_browser_state_show()
   
   #a63子樣板讀取&產生
   CALL adzp150_action_control()

END FUNCTION

#+ 產生狀態碼切換段落
PRIVATE FUNCTION adzp150_state_change()
   DEFINE ls_pre     STRING    #主table前置碼
   DEFINE ls_modid   STRING    #modid欄位名稱
   DEFINE ls_moddt   STRING    #moddt欄位名稱
   DEFINE ls_name    STRING  
   DEFINE ls_value   STRING  
    
   #先確定有沒有stus欄位
   IF g_properties.getValue("master_stus_isExist") <> "Y" THEN
      RETURN
   ELSE
      IF cl_null(g_properties.getValue("general_stus_cnt")) OR 
                 g_properties.getValue("general_stus_cnt") = 0 THEN
      END IF
   END IF
   
   #修改時異動modid,moddt
   LET ls_pre = g_properties.getValue("master_tbl_name")
   LET ls_pre = ls_pre.subString(1,ls_pre.getIndexOf('_',1)-1)
   LET ls_modid = ls_pre, 'modid'
   LET ls_moddt = ls_pre, 'moddt'
   IF adzp150_chk_field_exist(ls_modid,0) = 'Y' THEN
      LET ls_name  = 'mdl_field_modid'
      LET ls_value = ',',ls_modid
      CALL g_properties.addAttribute(ls_name,ls_value)
      LET ls_name  = 'mdl_var_modid_d'
      LET ls_value = ',',g_properties.getValue("master_var_title"),'.',ls_modid
      CALL g_properties.addAttribute(ls_name,ls_value)
      LET ls_name  = 'mdl_var_modid'
      LET ls_value = g_properties.getValue("master_var_title"),'.',ls_modid
      CALL g_properties.addAttribute(ls_name,ls_value)
   ELSE
      LET ls_name  = 'mdl_mark_modid'
      LET ls_value = "#該樣板不需此段落"
      CALL g_properties.addAttribute(ls_name,ls_value) 
   END IF
   IF adzp150_chk_field_exist(ls_moddt,0) = 'Y' THEN
      LET ls_name  = 'mdl_field_moddt'
      LET ls_value = ',',ls_moddt
      CALL g_properties.addAttribute(ls_name,ls_value)
      LET ls_name  = 'mdl_var_moddt_d'
      LET ls_value = ',',g_properties.getValue("master_var_title"),'.',ls_moddt
      CALL g_properties.addAttribute(ls_name,ls_value)
      LET ls_name  = 'mdl_var_moddt'
      LET ls_value = g_properties.getValue("master_var_title"),'.',ls_moddt
      CALL g_properties.addAttribute(ls_name,ls_value)
   ELSE
      LET ls_name  = 'mdl_mark_moddt'
      LET ls_value = "#該樣板不需此段落"
      CALL g_properties.addAttribute(ls_name,ls_value) 
   END IF
   
   IF g_properties.getValue("type_id") = "i01" OR 
      g_properties.getValue("type_id") = "i04" OR
      g_properties.getValue("type_id") = "i05" OR
      g_properties.getValue("type_id") = "i08" OR
      g_properties.getValue("type_id") = "t01" THEN
      CALL g_properties.addAttribute("general_state_change",adzp150_make_slice("a09"))
   END IF
   
   #IF g_properties.getValue("type_id") = "i02" OR
   #   g_properties.getValue("type_id") = "i07" OR
   #   g_properties.getValue("type_id") = "i12" THEN
   #   #CALL g_properties.addAttribute("general_state_change",adzp150_make_slice("a10"))
   #END IF
   
END FUNCTION

#+ 產生狀態碼action控制段落
PRIVATE FUNCTION adzp150_state_control()

   #CALL g_properties.addAttribute("general_state_control",adzp150_make_slice("a20"))

END FUNCTION

#+ 產生狀態碼圖片顯示段落
PRIVATE FUNCTION adzp150_state_show()

   CALL g_properties.addAttribute("general_state_show",adzp150_make_slice("a21"))

END FUNCTION

#+ 產生Browser狀態碼圖片顯示段落
PRIVATE FUNCTION adzp150_browser_state_show()

   CALL g_properties.addAttribute("browser_state_define",adzp150_make_slice("a24"))

END FUNCTION

#+ 產生狀態碼action控制段落
PRIVATE FUNCTION adzp150_state_action_control()
   DEFINE ls_name   STRING
   DEFINE ls_tmp    STRING
   
   #mdl_var_stus(單頭)
   IF g_properties.getValue("type_id") = "i01" OR 
      g_properties.getValue("type_id") = "i04" OR
      g_properties.getValue("type_id") = "i05" OR
      g_properties.getValue("type_id") = "i08" OR
      g_properties.getValue("type_id") = "t01" THEN
      LET ls_tmp = "master_var_title"
      LET ls_name = g_properties.getValue(ls_tmp),".",g_properties.getValue("master_field_stus")
   END IF
   
   #mdl_var_stus(單身)
   IF g_properties.getValue("type_id") = "i02" OR
      g_properties.getValue("type_id") = "t02" OR
      g_properties.getValue("type_id") = "i07" OR
      g_properties.getValue("type_id") = "i12" THEN
      LET ls_name = adzp150_find_page("stus",""), ".", g_properties.getValue("detail_field_stus")
   END IF

   CALL g_properties.addAttribute("mdl_var_stus",ls_name)
   
   #先產生需要的資訊
   CALL adzp150_stus_control()
   
   CALL g_properties.addAttribute("general_state_action_control",adzp150_make_slice("a22"))

END FUNCTION

#+ 準備stus相關資訊
PUBLIC FUNCTION adzp150_read_stus()
   DEFINE ps_template   STRING
   DEFINE ls_return     STRING 
   DEFINE lchannel_read base.Channel
   DEFINE ls_mdl        STRING
   DEFINE ls_read       STRING
   DEFINE ls_text       STRING
   DEFINE li_cnt        INTEGER
   DEFINE li_idx        INTEGER
   DEFINE ls_name       STRING
   DEFINE ls_value      STRING
   DEFINE ls_tmp        STRING
 
   #定義狀態資訊(mdl_stus,mdl_pic)
   LET li_cnt = g_properties.getValue("general_stus_cnt")
   FOR li_idx = 1 TO li_cnt
      LET ls_name  = "general_stus_type",li_idx USING "<<<"
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name  = "mdl_stus",li_idx USING "<<<"
      CALL g_properties.addAttribute(ls_name,ls_value)
      LET ls_tmp = ls_value
      LET ls_name  = "general_stus_pic",li_idx USING "<<<"
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name  = "mdl_pic",li_idx USING "<<<"
      CALL g_properties.addAttribute(ls_name,ls_value)
      IF ls_tmp = "W" OR ls_tmp = "D" THEN
         LET ls_name  = "mdl_stus_", ls_tmp
         CALL g_properties.addAttribute(ls_name,ls_value)
         LET ls_name  = "mdl_mark",li_idx USING "<<<"
         LET ls_value = "#"
         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF

   END FOR
   
   CALL g_properties.addAttribute("mdl_mdls",g_properties.getValue("general_stus_cnt")) 

   LET ls_tmp = g_properties.getValue("general_stus_scc")
   
   IF ls_tmp = '17' THEN
      LET ls_tmp = g_properties.getValue("general_stus_list")
      IF ls_tmp.getIndexOf('Y',1) THEN
         CALL g_properties.addAttribute("mdl_default_stus","Y") 
      ELSE
         CALL g_properties.addAttribute("mdl_default_mark","#該樣板不需此段落") 
      END IF
   ELSE
      LET ls_tmp = g_properties.getValue("general_stus_list")
      IF ls_tmp.getIndexOf('N',1) THEN
         CALL g_properties.addAttribute("mdl_default_stus","N") 
      ELSE
         CALL g_properties.addAttribute("mdl_default_mark","#該樣板不需此段落") 
      END IF
   END IF
   
END FUNCTION 

#+ 狀態碼相關資訊設定
PRIVATE FUNCTION  adzp150_stus_control()
   DEFINE li_idx       INTEGER
   DEFINE li_idx2      INTEGER
   DEFINE li_cnt       INTEGER
   DEFINE ls_slice     STRING
   DEFINE ls_name      STRING
   DEFINE ls_value     STRING
   DEFINE ls_pic       STRING
   
   #定義狀態資訊(mdl_stus,mdl_pic)
   LET li_cnt = g_properties.getValue("general_stus_cnt")
   
   FOR li_idx = 1 TO li_cnt
      LET ls_slice = ""
      FOR li_idx2 = 1 TO li_cnt
         
         LET ls_name  = "general_stus_pic",li_idx2 USING "<<<"
         LET ls_value = g_properties.getValue(ls_name)
         
         #相等代表目前所選的狀態為何
         IF li_idx = li_idx2 THEN
            LET ls_slice = ls_slice, 18 SPACES, "CALL cl_set_act_visible('",ls_value,"',FALSE)\n"
         ELSE
            LET ls_slice = ls_slice, 18 SPACES, "CALL cl_set_act_visible('",ls_value,"',TRUE)\n"
         END IF

      END FOR 

      #寫入控制段落
      LET ls_name  = "mdl_action_control",li_idx USING "<<<"
      LET ls_value = ls_slice
      CALL g_properties.addAttribute(ls_name,ls_value)
         
      #寫入狀態碼類別
      LET ls_name  = "general_stus_type",li_idx USING "<<<"
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name  = "mdl_stus_type",li_idx USING "<<<"
      CALL g_properties.addAttribute(ls_name,ls_value)
      
   END FOR
   
END FUNCTION 

#+ 產生簽核相關段落(t類程式使用)
PRIVATE FUNCTION adzp150_create_signature()
   DEFINE ls_prog       STRING
   DEFINE ls_list       STRING
   DEFINE ls_signature  STRING
   
   LET ls_prog = g_properties.getValue("app_id")
   IF ls_prog.subString(4,4) <> 't' THEN
      CALL g_properties.addAttribute("general_signature_mark","#該樣板不需此段落:") 
      RETURN
   END IF
   
   LET ls_list = g_properties.getValue("general_stus_list")
   IF ls_list.getIndexOf('W',1) = 0 THEN
      CALL g_properties.addAttribute("general_signature_mark","#該樣板不需此段落:") 
      RETURN
   END IF
   
   LET ls_signature = adzp150_make_slice("a32")
   CALL g_properties.addAttribute("general_menu_signature",ls_signature) 

   CALL g_properties.addAttribute("mdl_num",'2') 
   LET ls_signature = adzp150_make_slice("a32")
   CALL g_properties.addAttribute("general_menu_signature2",ls_signature) 
   
   #LET ls_signature = adzp150_make_slice("a33")
   #CALL g_properties.addAttribute("mdl_a33",ls_signature) 
   #LET ls_signature = adzp150_make_slice("a34")
   #CALL g_properties.addAttribute("mdl_a34",ls_signature) 
   #LET ls_signature = adzp150_make_slice("a35")
   #CALL g_properties.addAttribute("mdl_a35",ls_signature) 
   
   LET ls_signature = adzp150_make_slice("a36")
   CALL g_properties.addAttribute("mdl_a36",ls_signature) 
   #LET ls_signature = adzp150_make_slice("a37")
   #CALL g_properties.addAttribute("mdl_a37",ls_signature) 
   
   IF cl_null(g_properties.getValue("detail_var_title")) THEN   #表示為單檔
      CALL g_properties.addAttribute("mdl_mark","#該樣板不需此段落")
   END IF

   #170322-00081 ---start---
   IF NOT cl_null(g_properties.getValue("detail_var_title")) THEN   #表示為雙檔
      CALL g_properties.addAttribute("mdl_not_signal_mark","#該樣板不需此段落")
   END IF
   
   #170322-00081 --- end ---
   
   LET ls_signature = adzp150_make_slice("a39"),'\n',adzp150_make_slice("a40")
   CALL g_properties.addAttribute("general_func_signature",ls_signature) 
   
   CALL g_properties.addAttribute("mdl_mark","") 
   
   LET ls_signature = adzp150_make_slice("a56")
   CALL g_properties.addAttribute("general_signature_chk",ls_signature) 
   
END FUNCTION

#+ 寫入控管Action的段落
PRIVATE FUNCTION adzp150_action_control()
   DEFINE ls_name      STRING
   DEFINE ls_value     STRING
   
   #先確定是否需要(只有scc=13需要)
   LET ls_name  = "general_stus_scc"
   LET ls_value = g_properties.getValue(ls_name)
   IF ls_value <> '13' OR cl_null(ls_value) OR
      #目前只做t01樣板
      g_properties.getValue("type_id") <> "t01" THEN
      RETURN
   END IF
   
   #先準備變數
   LET ls_name  = "mdl_var_stus"
   LET ls_value = g_properties.getValue("master_var_title"),".",g_properties.getValue("master_field_stus")
   CALL g_properties.addAttribute(ls_name,ls_value) 
   
   #寫入add-point
   LET g_dzbb.prog_name   = g_properties.getValue("app_id")
   LET g_dzbb.point_name  = "set_act_no_visible.set_act_no_visible"
   LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
   LET g_dzbb.addpoint    = adzp150_make_slice("a63")
   CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))
               
END FUNCTION

