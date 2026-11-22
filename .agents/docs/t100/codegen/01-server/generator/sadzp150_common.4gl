#adzp150 副程式 - 公用欄位相關處理
# Memo : 190412-00022#1 by 00824   資料所有者與所有部門改為呼叫FUNCTION取得


GLOBALS "../4gl/adzp150_global.inc"   #180108-00012
GLOBALS
   DEFINE g_properties   om.SaxAttributes
   CONSTANT li_space = 3
END GLOBALS

PUBLIC FUNCTION adzp150_create_common()
   DEFINE ls_m_tbl          STRING
   DEFINE ls_d_tbl          STRING
   DEFINE ls_slice          STRING
   DEFINE ls_common         STRING
   DEFINE ls_exist          STRING
   DEFINE ls_tbl_common     DYNAMIC ARRAY OF RECORD
          name  STRING,
          exist STRING
          END RECORD
   DEFINE ls_common_name    DYNAMIC ARRAY OF STRING
   DEFINE li_idx            INTEGER
   DEFINE li_idx2           INTEGER
   DEFINE ls_tmp            STRING
   DEFINE ls_name           STRING
   DEFINE ls_value          STRING
   DEFINE ls_tbl_pre        STRING
   DEFINE ls_pages          STRING
   DEFINE lst_token         base.StringTokenizer
   DEFINE ls_token          STRING
   DEFINE ls_title_pre      STRING
   
   LET ls_common_name[1]  = "ownid"
   LET ls_common_name[2]  = "owndp"
   LET ls_common_name[3]  = "crtid"
   LET ls_common_name[4]  = "crtdp"
   LET ls_common_name[5]  = "crtdt"
   LET ls_common_name[6]  = "modid"
   LET ls_common_name[7]  = "moddt"
   LET ls_common_name[8]  = "cnfid"
   LET ls_common_name[9]  = "cnfdt"
   LET ls_common_name[10] = "pstid"
   LET ls_common_name[11] = "pstdt"
   LET ls_common_name[12] = "stus"
   
   #單頭段處理
   FOR li_idx = 1 TO g_properties.getValue("master_tbl_cnt")
   
      INITIALIZE ls_tbl_common TO NULL
      
      LET ls_tmp =  adzp150_create_name(li_idx, "master_tbl_name_by_tbl", "<<<") 
      LET ls_tmp = g_properties.getValue(ls_tmp)
      LET ls_tbl_pre = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      
      #tbl name
      LET ls_name = "mdl_tbl_name"
      LET ls_value = ls_tmp
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      LET ls_exist = "N"
      #組合各table公用欄位名稱並判斷是否存在
      FOR li_idx2 = 1 TO ls_common_name.getLength()
         LET ls_tbl_common[li_idx2].name = ls_tbl_pre, ls_common_name[li_idx2]
         LET ls_tbl_common[li_idx2].exist = adzp150_chk_field_exist(ls_tbl_common[li_idx2].name,0)
         IF ls_tbl_common[li_idx2].exist = "Y" THEN
            LET ls_exist = "Y"
         END IF

         #ownid跟dept專用mark
         IF ls_tbl_common[li_idx2].exist = "N" AND li_idx = 1 THEN
            LET ls_name = "master_mark_", ls_common_name[li_idx2]
            CALL g_properties.addAttribute(ls_name, "#該樣板不需此段落")
         END IF
      END FOR
      
      IF ls_exist = "Y" THEN
      
         #將公用欄位名稱加入mdl變數中
         LET ls_tmp = adzp150_create_name(li_idx, "master_var_title", "<<<") 
         LET ls_title_pre = g_properties.getValue(ls_tmp),"."
         
         FOR li_idx2 = 1 TO ls_tbl_common.getLength()
            #前置字元
            LET ls_name = adzp150_create_name(li_idx2, "mdl_var_title", "&&")
            LET ls_value = ls_title_pre
            CALL g_properties.addAttribute(ls_name, ls_value)
            #不存在則mark
            IF ls_tbl_common[li_idx2].exist = "N" THEN
               LET ls_value = "#該樣板不需此段落"
            ELSE
               LET ls_value = ""
            END IF
            LET ls_name = adzp150_create_name(li_idx2, "mdl_mark", "&&")
            CALL g_properties.addAttribute(ls_name, ls_value)
            #欄位名稱
            LET ls_name = adzp150_create_name(li_idx2, "mdl_field", "&&")
            LET ls_value = ls_tbl_common[li_idx2].name
            CALL g_properties.addAttribute(ls_name, ls_value)
            #變數名稱
            LET ls_name = adzp150_create_name(li_idx2, "mdl_var", "&&")
            LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
            CALL g_properties.addAttribute(ls_name, ls_value)
            IF li_idx2 = 1 THEN
               #ownid
               LET ls_name = "master_var_ownid"
               CALL g_properties.addAttribute(ls_name, ls_value)
            END IF
            IF li_idx2 = 2 THEN
               #owndp 
               LET ls_name = "master_var_owndp"
               CALL g_properties.addAttribute(ls_name, ls_value)
            END IF
            #180201-00054 ---start---
            IF li_idx2 = 3 THEN
               #crtid
               LET ls_name = "master_var_crtid"
               CALL g_properties.addAttribute(ls_name, ls_value)
            END IF
            IF li_idx2 = 4 THEN
               #crtdp
               LET ls_name = "master_var_crtdp"
               CALL g_properties.addAttribute(ls_name, ls_value)
            END IF
            #180201-00054 --- end ---
            #變數名稱(pre)
            LET ls_name = adzp150_create_name(li_idx2, "mdl_pre", "&&")
            LET ls_value = ls_tbl_common[li_idx2].name
            CALL g_properties.addAttribute(ls_name, ls_value)
         END FOR
         
         CALL g_properties.addAttribute("mdl_var_title",ls_title_pre)
         CALL g_properties.addAttribute("mdl_loc", "master")
      
         #公用欄位開窗
         LET ls_slice = g_properties.getValue("master_commons_ctrlp")
         LET ls_slice = ls_slice, adzp150_make_slice("a11")
         CALL g_properties.addAttribute("master_commons_ctrlp",ls_slice)
         
         #公用欄位帶值
         LET ls_slice = g_properties.getValue("master_commons_ref")
         LET ls_slice = ls_slice, adzp150_make_slice("a12")
         CALL g_properties.addAttribute("master_commons_ref",ls_slice)
         
         #公用欄位複製給值
         LET ls_slice = g_properties.getValue("master_repro_define")
         LET ls_slice = ls_slice, adzp150_make_slice("a13")
         CALL g_properties.addAttribute("master_repro_define",ls_slice)
         
         #公用欄位新增給值 
         LET ls_slice = g_properties.getValue("master_insert_define")
         LET ls_slice = ls_slice, adzp150_make_slice("a14")
         CALL g_properties.addAttribute("master_insert_define",ls_slice)
         
         #公用欄位owner定義
         #LET ls_slice = g_properties.getValue("master_owner_define")
         #LET ls_slice = ls_slice, adzp150_common_create_owner(li_idx)
         #CALL g_properties.addAttribute("master_owner_define",ls_slice)
         
         #公用欄位更改時給值
         LET ls_slice = g_properties.getValue("master_upder_define")
         LET ls_slice = ls_slice, adzp150_common_create_upder(li_idx)
         CALL g_properties.addAttribute("master_upder_define",ls_slice)
         
      END IF
      
   END FOR
   
   #單身段處理
   FOR li_idx = 1 TO g_properties.getValue("detail_tbl_cnt")
   
      INITIALIZE ls_tbl_common TO NULL
      
      LET ls_tmp =  adzp150_create_name(li_idx, "detail_tbl_name_by_tbl", "<<<") 
      LET ls_tmp = g_properties.getValue(ls_tmp)
      LET ls_tbl_pre = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      
      LET ls_exist = "N"
      #組合各table公用欄位名稱並判斷是否存在
      FOR li_idx2 = 1 TO ls_common_name.getLength()
         LET ls_tbl_common[li_idx2].name = ls_tbl_pre, ls_common_name[li_idx2]
         LET ls_tbl_common[li_idx2].exist = adzp150_chk_field_exist(ls_tbl_common[li_idx2].name,li_idx)
         IF ls_tbl_common[li_idx2].exist = "Y" THEN
            LET ls_exist = "Y"
         END IF

         #161109-00070 #1 ---add start---
         #ownid跟dept專用mark
         IF ls_tbl_common[li_idx2].exist = "N" AND li_idx = 1 THEN
            LET ls_name = "detail_mark_", ls_common_name[li_idx2]
            CALL g_properties.addAttribute(ls_name, "#該樣板不需此段落")
         END IF
         #161109-00070 #1 --- add end ---
      END FOR

      IF ls_exist = "Y" THEN

         #將公用欄位名稱加入mdl變數中
         FOR li_idx2 = 1 TO ls_tbl_common.getLength()
            #不存在則mark
            IF ls_tbl_common[li_idx2].exist = "N" THEN
               LET ls_value = "#該樣板不需此段落"
            ELSE
               LET ls_value = ""
            END IF
            LET ls_name = adzp150_create_name(li_idx2, "mdl_mark", "&&")
            CALL g_properties.addAttribute(ls_name, ls_value)
            #前置字元
            LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_pages", "<<<") 
            LET ls_tmp = g_properties.getValue(ls_tmp)
            LET ls_title_pre = adzp150_find_page(ls_tbl_common[li_idx2].name,ls_tmp),'.'
            LET ls_name = adzp150_create_name(li_idx2, "mdl_var_title", "&&")
            LET ls_value = ls_title_pre
            CALL g_properties.addAttribute(ls_name, ls_value)
            #欄位名稱
            LET ls_name = adzp150_create_name(li_idx2, "mdl_field", "&&")
            LET ls_value = ls_tbl_common[li_idx2].name
            CALL g_properties.addAttribute(ls_name, ls_value)
            #變數名稱
            LET ls_name = adzp150_create_name(li_idx2, "mdl_var", "&&")
            LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
            LET ls_value = cl_replace_str(ls_value,'[l_ac]','[1]')
            CALL g_properties.addAttribute(ls_name, ls_value)
            #161109-00070 #1 ---add start---
            IF li_idx = 1 AND li_idx2 = 1 THEN   #170804-00012 add li_idx
               #ownid
               LET ls_name = "detail_var_ownid"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               LET ls_value = cl_replace_str(ls_value,'[l_ac]','[g_detail_idx]')
               CALL g_properties.addAttribute(ls_name, ls_value)

               #190412-00022 ---start---
               LET ls_name = "detail_var_ownid_v2"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #190412-00022 --- end ---

               #171006-00020 ---start---
               LET ls_name = "detail_field_ownid"
               LET ls_value = ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #171006-00020 --- end ---
            END IF

            #180201-00054 ---modify start---
            IF li_idx = 1 AND li_idx2 = 2 THEN   #170316-00030 add li_idx
               #owndp
               LET ls_name = "detail_var_owndp"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               LET ls_value = cl_replace_str(ls_value,'[l_ac]','[g_detail_idx]')
               CALL g_properties.addAttribute(ls_name, ls_value)
               
               #190412-00022 ---start---
               LET ls_name = "detail_var_owndp_v2"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #190412-00022 --- end ---

               #171006-00020 ---start---
               LET ls_name = "detail_field_owndp"
               LET ls_value = ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #171006-00020 --- end ---
           END IF

            IF li_idx = 1 AND li_idx2 = 3 THEN   #170316-00030 add li_idx
               #crtid
               LET ls_name = "detail_var_crtid"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               LET ls_value = cl_replace_str(ls_value,'[l_ac]','[g_detail_idx]')
               CALL g_properties.addAttribute(ls_name, ls_value)
           
               #190412-00022 ---start---
               LET ls_name = "detail_var_crtid_v2"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #190412-00022 --- end ---

               #171006-00020 ---start---
               LET ls_name = "detail_field_crtid"
               LET ls_value = ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #171006-00020 --- end ---
           END IF

            IF li_idx = 1 AND li_idx2 = 4 THEN   #170316-00030 add li_idx
               #crtdp
               LET ls_name = "detail_var_crtdp"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               LET ls_value = cl_replace_str(ls_value,'[l_ac]','[g_detail_idx]')
               CALL g_properties.addAttribute(ls_name, ls_value)
           
               #190412-00022 ---start---
               LET ls_name = "detail_var_crtdp_v2"
               LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #190412-00022 --- end ---

               #171006-00020 ---start---
               LET ls_name = "detail_field_crtdp"
               LET ls_value = ls_tbl_common[li_idx2].name
               CALL g_properties.addAttribute(ls_name, ls_value)
               #171006-00020 --- end ---

               #for單檔多欄型態樣板(i02, i07)
               #若有設定公用欄位，取資料時須加上權限控管條件
               IF ls_tbl_common[li_idx2].exist = "Y" THEN
                  LET ls_slice = adzp150_make_slice("a68")
                  CALL g_properties.addAttribute("detail_ownid_condition",ls_slice)

                  #是否顯示訊息
                  LET ls_name = "detail_ownid_modify_msg"
                  LET ls_value = "IF g_action_choice = 'modify' OR g_action_choice = 'modify_detail' THEN\n",
                                 #170316-00030 ---modify start---
                                 6 SPACES,"INITIALIZE g_errparam TO NULL\n",
                                 6 SPACES,"LET g_errparam.extend = ''\n",
                                 6 SPACES,"LET g_errparam.code   = 'lib-00419'\n",
                                 6 SPACES,"LET g_errparam.popup  = FALSE\n",   #170522-00054 改為FALSE
                                 6 SPACES,"CALL cl_err()\n",
                                 3 SPACES,"END IF"
                                 #170316-00030 --- modify end ---
                  CALL g_properties.addAttribute(ls_name, ls_value)
            #170316-00030 ---mark start---
            #  ELSE
            #     LET ls_name = "detail_ownid_modify_msg"
            #     LET ls_value = ""
            #     CALL g_properties.addAttribute(ls_name, ls_value)
            #170316-00030 --- mark end ---
               END IF
            END IF
            ##180201-00054 --- modify end ---
            #161109-00070 #1 --- add end ---
            #變數名稱(pre)
            LET ls_name = adzp150_create_name(li_idx2, "mdl_pre", "&&")
            LET ls_value = ls_tbl_common[li_idx2].name
            CALL g_properties.addAttribute(ls_name, ls_value)
         END FOR
         
         LET ls_value = adzp150_create_name(li_idx, "detail", "<<<")
         CALL g_properties.addAttribute("mdl_loc", ls_value)
      
         #公用欄位開窗
         LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_name_by_tbl", "<<<") 
         LET ls_slice = adzp150_make_slice("a11")
         LET ls_tmp = adzp150_create_name(li_idx, "detail_commons_ctrlp", "<<<") 
         CALL g_properties.addAttribute(ls_tmp,ls_slice)
         
         #替換[1]為[l_ac]
         FOR li_idx2 = 1 TO ls_tbl_common.getLength()
            #前置字元
            LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_pages", "<<<") 
            LET ls_tmp = g_properties.getValue(ls_tmp)
            LET ls_title_pre = adzp150_find_page(ls_tbl_common[li_idx2].name,ls_tmp),'.'
            #變數名稱
            LET ls_name = adzp150_create_name(li_idx2, "mdl_var", "&&")
            LET ls_value = ls_title_pre, ls_tbl_common[li_idx2].name
            CALL g_properties.addAttribute(ls_name, ls_value)
         END FOR
         
         #公用欄位帶值
         LET ls_slice = adzp150_make_slice("a12")
         LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_pages", "<<<") 
         LET ls_tmp = g_properties.getValue(ls_tmp)
         LET ls_pages = adzp150_find_page_num(ls_tbl_common[1].name,ls_tmp)
         LET ls_tmp = adzp150_create_name(ls_pages, "detail_commons_ref", "<<<") 
         CALL g_properties.addAttribute(ls_tmp,ls_slice)
         
         #公用欄位複製給值
         LET ls_slice = g_properties.getValue("detail_repro_define")
         LET ls_slice = ls_slice, adzp150_make_slice("a13")
         CALL g_properties.addAttribute("detail_repro_define",ls_slice)
         
         #公用欄位新增給值
         LET ls_slice = adzp150_make_slice("a14")
         LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_pages", "<<<") 
         LET ls_pages = g_properties.getValue(ls_tmp)
         
         #開始解析那些page屬於該table
         LET lst_token = base.StringTokenizer.create(ls_pages, ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_tmp = adzp150_create_name(ls_token, "detail_insert_define", "<<<") 
            CALL g_properties.addAttribute(ls_tmp,ls_slice)
         END WHILE
         
         #公用欄位owner定義
         #LET ls_slice = g_properties.getValue("detail_owner_define")
         #LET ls_slice = ls_slice, adzp150_common_create_owner(li_idx)
         #CALL g_properties.addAttribute("detail_owner_define",ls_slice)
         
         #公用欄位更改時給值 
         LET ls_slice = adzp150_common_create_upder(li_idx)
         LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_pages", "<<<") 
         LET ls_pages = g_properties.getValue(ls_tmp)
         
         #開始解析那些page屬於該table
         LET lst_token = base.StringTokenizer.create(ls_pages, ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_tmp = adzp150_create_name(ls_token, "detail_upder_define", "<<<") 
            CALL g_properties.addAttribute(ls_tmp,ls_slice)
         END WHILE
         
      END IF  

   END FOR   
    
END FUNCTION

#+ 公用欄位owner定義
PRIVATE FUNCTION adzp150_common_create_owner(ps_tbl)
   DEFINE ps_tbl      STRING
   DEFINE ls_return   STRING
   
   LET ls_return = "LET g_data_owner = ",g_properties.getValue("mdl_var03"),"\n",
                   "LET g_data_group = ",g_properties.getValue("mdl_var04"),"\n" 

   RETURN ls_return
   
END FUNCTION

#+ 公用欄位更改時給值
PRIVATE FUNCTION adzp150_common_create_upder(ps_tbl)
   DEFINE ps_tbl      STRING
   DEFINE ls_return   STRING
   
   LET ls_return = g_properties.getValue("mdl_mark06"),
                   "LET ",g_properties.getValue("mdl_var06")," = g_user \n",
                   g_properties.getValue("mdl_mark07"),
                   "LET ",g_properties.getValue("mdl_var07")," = cl_get_current()\n",
                   g_properties.getValue("mdl_mark06"),
                   "LET ",g_properties.getValue("mdl_var06"),"_desc = cl_get_username(",g_properties.getValue("mdl_var06"),")"           
   RETURN ls_return
   
END FUNCTION

#+ 檢查該欄位是否存在於畫面上
PUBLIC FUNCTION adzp150_chk_field_exist(ps_field,pi_idx)
   DEFINE ps_field        STRING  
   DEFINE pi_idx          INTEGER  
   DEFINE ls_return       STRING  #Y-有, N-無
   DEFINE ls_page_list    STRING  #先抓出這個table涵蓋的page
   DEFINE ls_field_list   STRING  #再抓出這個table出現在畫面上的欄位
   DEFINE ls_field        STRING
   DEFINE ls_field2       STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_name         STRING
   
   IF pi_idx = 0 THEN
      #單頭
      LET ls_field_list = g_properties.getValue("general_form_field")
   ELSE
      #單身(判斷有哪些page)
      LET ls_name = adzp150_create_name(pi_idx, "detail_tbl_pages", "<<<")
      LET ls_page_list = g_properties.getValue(ls_name)
      LET lst_token = base.StringTokenizer.create(ls_page_list, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken() 
         
         #尋找對應page的欄位
         LET ls_name = "detail_field_loc", ls_token
         LET ls_field_list = ls_field_list, g_properties.getValue(ls_name), ','
         
      END WHILE
   END IF
   
   LET ls_field = ps_field,","
   LET ls_field2 = ps_field,"("
   
   IF ls_field_list.getIndexOf(ls_field,1) OR 
      ls_field_list.getIndexOf(ls_field2,1) THEN
      LET ls_return = 'Y'
   ELSE
      LET ls_return = 'N'
   END IF
   
   RETURN ls_return

END FUNCTION
