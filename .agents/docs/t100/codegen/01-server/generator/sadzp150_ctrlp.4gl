#adzp150 副程式 - 開窗相關處理
#250212-00006#1 2025/02/12 by 20992 配合項標中心對於設計器優化需求，增加開窗校驗帶值功能

IMPORT os
SCHEMA ds

GLOBALS "../4gl/adzp150_global.inc"   #180108-00012
GLOBALS
   DEFINE g_properties   om.SaxAttributes
END GLOBALS

#+ 生成controlp元件
PUBLIC FUNCTION adzp150_create_controlp(p_node,pi_lv,p_loc)
   DEFINE p_node        om.DomNode
   DEFINE pi_lv         LIKE type_t.num10 
   DEFINE p_loc         STRING
   DEFINE ls_return     STRING 
   DEFINE lchannel_read base.Channel
   DEFINE ls_mdl        STRING
   DEFINE ls_read       STRING
   DEFINE ls_text       STRING
   
   #若無開窗代號則直接返回
   IF cl_null(p_node.getAttribute("form")) THEN
      RETURN ""
   END IF
   
   #取得所需參數
   CALL adzp150_get_para(p_node)
    
   #讀取小樣板
   IF p_loc = "construct" OR p_loc = "query" THEN
      LET ls_mdl = "a08"
   ELSE
      LET ls_mdl = "a07"
   END IF
   
   #讀取子樣板
   LET ls_return = adzp150_make_slice(ls_mdl)

   LET ls_return = adzp150_add_ctrlp_args(p_node, ls_return, p_loc) #250212-00006#1 add
   
   RETURN ls_return

END FUNCTION 

#取得必要參數
PRIVATE FUNCTION adzp150_get_para(p_node)
   DEFINE p_node          om.DomNode
   DEFINE ls_fields       DYNAMIC ARRAY OF LIKE type_t.chr500
   DEFINE ls_vars         DYNAMIC ARRAY OF LIKE type_t.chr500
   DEFINE ls_descs        DYNAMIC ARRAY OF LIKE type_t.chr500
   DEFINE ls_arg_descs    DYNAMIC ARRAY OF LIKE type_t.chr500
   DEFINE ls_var_title    STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_sql          STRING
   DEFINE ls_tmp          STRING
   DEFINE li_cnt          INTEGER
   
   #取得開窗代碼
   CALL g_properties.addAttribute("mdl_form",p_node.getAttribute("form"))
   
   #取得變數title
   IF g_properties.getValue("location") = "head" THEN
      LET ls_var_title = g_properties.getValue("master_var_title"),"."
   ELSE
      IF g_properties.getValue("general_current_page") = "1" THEN
         LET ls_tmp = "detail_var_title"
      ELSE
         LET ls_tmp = "detail_var_title",g_properties.getValue("general_current_page")
      END IF
      LET ls_var_title = g_properties.getValue(ls_tmp),"[l_ac]."
   END IF
   
   #撈出arg的說明
   LET ls_sql = " SELECT dzcb004 ",
                " FROM   dzcb_t ",
                " WHERE  dzcb001 = '",p_node.getAttribute("form"),"' ",
                " AND    dzcb003 IS NOT NULL ",
                " ORDER BY dzcb003"
              
   PREPARE dzzb_desc FROM ls_sql
   DECLARE dzzb_desccur CURSOR FOR dzzb_desc
   
   LET li_idx = 1
   FOREACH dzzb_desccur INTO ls_arg_descs[li_idx]
      LET ls_tmp = "mdl_arg_desc", li_idx USING "<<<"
      CALL g_properties.addAttribute(ls_tmp, ls_arg_descs[li_idx])
      LET li_idx = li_idx + 1
   END FOREACH
   
   #確定共有幾個arg
   LET li_cnt = ls_arg_descs.getLength() - 1
   CALL g_properties.addAttribute("mdl_mdls2",li_cnt)
   
   #第一個欄位名稱
   #LET ls_fields[1] = g_properties.getValue("column")
   
   #2-N欄位名稱
   LET ls_sql = " SELECT dzcc003,dzeb003 ",
                " FROM   dzcc_t ",
                " LEFT OUTER JOIN dzeb_t ON dzcc003 = dzeb002 ",
                " WHERE  dzcc001 = '",p_node.getAttribute("form"),"' ",
                " AND    dzcc005 = 'Y' "
              
   PREPARE dzzc_field FROM ls_sql
   DECLARE dzzc_fieldcur CURSOR FOR dzzc_field
   
   LET li_idx = 1
   FOREACH dzzc_fieldcur INTO ls_fields[li_idx],ls_descs[li_idx]
      LET li_idx = li_idx + 1
   END FOREACH
   CALL ls_fields.deleteElement(ls_fields.getLength())
   
   #第一個欄位名稱
   LET ls_fields[1] = g_properties.getValue("column")
   
   #組合1-N變數名稱
   FOR li_idx = 1 TO ls_fields.getLength()
      LET ls_vars[li_idx] = ls_var_title,ls_fields[li_idx]
   END FOR
   
   #加入g_properties(field,var,desc)
   FOR li_idx = 1 TO ls_fields.getLength()
      LET ls_tmp = "mdl_field", li_idx USING "<<<"
      CALL g_properties.addAttribute(ls_tmp, ls_fields[li_idx])
      LET ls_tmp = "mdl_var", li_idx USING "<<<"
      CALL g_properties.addAttribute(ls_tmp, ls_vars[li_idx])
      LET ls_tmp = "mdl_desc", li_idx USING "<<<"
      CALL g_properties.addAttribute(ls_tmp, ls_descs[li_idx])
   END FOR
   
   #確定回傳幾個欄位
   CALL g_properties.addAttribute("mdl_mdls",ls_vars.getLength())
   
END FUNCTION

#250212-00006#1 add --- start ---
#+ 增加開窗參數資料
PUBLIC FUNCTION adzp150_add_ctrlp_args(pn_node, l_return, p_loc)
   DEFINE pn_node       om.DomNode
   DEFINE l_return        STRING
   DEFINE p_loc         STRING
   DEFINE lc_dzcy DYNAMIC ARRAY OF RECORD
      ls_dzcy005    LIKE dzcy_t.dzcy005,  #參數順序(PK)
      ls_dzcy007    LIKE dzcy_t.dzcy007   #參數值使用控件
   END RECORD
   DEFINE ls_sql       STRING
   DEFINE l_ac         INTEGER
   DEFINE l_ck         INTEGER
   DEFINE ls_type       STRING
   DEFINE ls_arg       STRING
   DEFINE ls_str       STRING
   DEFINE ls_str1      STRING

   IF p_loc = "construct" OR p_loc = "query" THEN
      LET ls_type = "2" #2:查詢時開窗
   ELSE
      LET ls_type = "1" #1:編輯時開窗
   END IF
   
   #先判斷是否有客製，有客製以客製處理
   LET ls_sql = " SELECT COUNT(*) FROM dzcy_t WHERE dzcy003 = '", ls_type, "' AND  dzcy001 = '", g_properties.getValue("app_id"),"'",
                     " AND dzcy004 = '", pn_node.getAttribute("form"),"' AND dzcy002 LIKE '%", g_properties.getValue("column"),"' ",
                     " AND dzcy010 = 'c' " 
   PREPARE adzp150_create_controlp_dzcy_sel FROM ls_sql
   EXECUTE adzp150_create_controlp_dzcy_sel INTO l_ck

   LET ls_sql = " SELECT dzcy005, dzcy007 FROM dzcy_t WHERE dzcy003 = '", ls_type, "' AND  dzcy001 = '", g_properties.getValue("app_id"),"'",
                  " AND dzcy004 = '", pn_node.getAttribute("form"),"' AND dzcy002 LIKE '%", g_properties.getValue("column"),"' "
   IF l_ck > 0 THEN
      LET ls_sql = ls_sql, " AND dzcy010 = 'c' "
   ELSE
      LET ls_sql = ls_sql, " AND dzcy010 = 's'  "
   END IF
   IF ls_type = "2" THEN
      LET ls_sql = ls_sql, " ORDER BY dzcy005 DESC "
   ELSE
      LET ls_sql = ls_sql, " ORDER BY dzcy005 "
   END IF
   PREPARE adzp150_create_controlp_dzcy_pre FROM ls_sql
   DECLARE adzp150_create_controlp_dzcy_cur CURSOR FOR adzp150_create_controlp_dzcy_pre
   CALL lc_dzcy.clear()
   LET l_ac = 1
   FOREACH adzp150_create_controlp_dzcy_cur INTO lc_dzcy[l_ac].ls_dzcy005, lc_dzcy[l_ac].ls_dzcy007
      #傳入參數轉換
      LET ls_arg = lc_dzcy[l_ac].ls_dzcy007
      #【:】開頭才需要處理轉換，沒有 : 就直接使用
      IF ls_arg.subString(1,1) = ":" THEN
         CASE   
            #指定字符採用正面表列處理
            WHEN ls_arg = ":TODAY"
               LET ls_arg = "g_today"
            WHEN ls_arg = ":LANG"
               LET ls_arg = "g_lang"
            WHEN ls_arg = ":DLANG"
               LET ls_arg = "g_dlang"
            WHEN ls_arg = ":ENT"
               LET ls_arg = "g_enterprise"
            WHEN ls_arg = ":SITE"
               LET ls_arg = "g_site"
            WHEN ls_arg = ":DEPT"
               LET ls_arg = "g_dept"
            WHEN ls_arg = ":USER"
               LET ls_arg = "g_user"
            WHEN ls_arg = ":CURRENT"
               LET ls_arg = "cl_get_current()"
            WHEN ls_arg = ":LEGAL"
               LET ls_arg = "g_legal"
            OTHERWISE
               #其他轉換 :itzs_t.itzs003 -> g_xmda_m.xmda004
               CASE g_properties.getValue("location") 
                  WHEN "head"  #單頭                  
                     LET ls_arg = "g_", ls_arg.subString(2, ls_arg.getIndexOf(".", 1) - 3), "_m", 
                        ls_arg.subString(ls_arg.getIndexOf(".", 1), ls_arg.getLength())
                  WHEN "body"  #單身
                     LET ls_arg = "g_", ls_arg.subString(2, ls_arg.getIndexOf(".", 1) - 3), "_d", 
                        ls_arg.subString(ls_arg.getIndexOf(".", 1), ls_arg.getLength())
               END CASE
         END CASE
      END IF
      
      #調整取代指定字串
      IF ls_type = "2" THEN
         LET ls_str = "            LET g_qryparam.arg", lc_dzcy[l_ac].ls_dzcy005 USING "<<<", " = " #組合轉換參數字串
         LET ls_str1 = "LET g_qryparam.reqry = " #取出比對用字串
         LET l_return = l_return.subString(1, l_return.getIndexOf("\n", l_return.getIndexOf(ls_str1, 1))), 
                  ls_str, ls_arg,  l_return.subString(l_return.getIndexOf("\n", l_return.getIndexOf(ls_str1, 1)+1), l_return.getLength())
      ELSE
         LET ls_str = "LET g_qryparam.arg", l_ac USING "<<<", " = " #取出待轉換參數字串
         LET ls_str1 = l_return.subString(l_return.getIndexOf(ls_str, 1) -1, l_return.getIndexOf("\n", l_return.getIndexOf(ls_str, 1)) - 1)
         LET l_return = l_return.subString(1, l_return.getIndexOf(ls_str, 1) -1), ls_str, ls_arg, 
            l_return.subString(l_return.getIndexOf("\n", l_return.getIndexOf(ls_str1, 1) + 1), l_return.getLength())
         
      END IF
      
      LET l_ac = l_ac + 1
   END FOREACH
   CALL lc_dzcy.deleteElement(l_ac)
   
   RETURN l_return

END FUNCTION 
#250212-00006#1 add --- end ---

