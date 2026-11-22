#adzp150 副程式 - 多table處理


#160617         第三階單身子表調整
#170424-00031#1 a06格式調整
#190730-00003#4 F提效調整(增加單身pre字串)
#PGS-02667   ps_keys型態轉換
#200804-00006#1 2020/09/09 by 00824   Postgre環境調整的部分協助補過單(搜尋PGS-02667字樣)
#250527-00018#1 2025/05/27 by 20992   修正產生器判斷所屬表格問題

SCHEMA ds

GLOBALS "../4gl/adzp150_global.inc"   #180108-00012
GLOBALS
   DEFINE g_properties   om.SaxAttributes
   DEFINE g_update       STRING
END GLOBALS

DEFINE g_value       STRING
DEFINE g_type        STRING

DEFINE g_master_table_info   DYNAMIC ARRAY OF RECORD
         idx         INTEGER,
         id          STRING,
         pk          STRING,
         forupd_sql  STRING,
         fill_sql    STRING,
         ent         STRING,
         site        STRING
       END RECORD

DEFINE g_table_group     DYNAMIC ARRAY OF RECORD
         id          STRING,
         page        STRING,
         page2       STRING,
         page_insert STRING,
         page_update STRING,
         page_delete STRING
       END RECORD
       
DEFINE g_detail_table_info   DYNAMIC ARRAY OF RECORD
         idx         INTEGER,
         id          STRING,
         page        STRING,
         first_page  STRING,
         first_key   STRING,
         fk          STRING,
         mapping_fk  STRING,
         pk          STRING,
         var_fk      STRING,
         var_fk_t    STRING,
         var_pk      STRING,
         var_pk_t    STRING,
         linked      STRING,
         add_wc      STRING,
         curr_idx    STRING,
         forupd_sql  STRING,
         fill_sql    STRING,
         type        STRING, #m or d
         master      STRING, #上層
         ent         STRING,
         site        STRING,
         site_pk     STRING
       END RECORD
       
DEFINE g_detail_page_info   DYNAMIC ARRAY OF RECORD
         id            STRING,
         fk            STRING,
         pk            STRING,
         var_fk        STRING,
         var_pk        STRING,
         var_fk_t      STRING,
         var_pk_t      STRING,
         sync_ins      STRING,
         sync_upd      STRING,
         sync_del_pre  STRING,
         sync_del_post STRING,
         sql_wc        STRING,
         sql_wc_t      STRING,
         first_page    STRING,
         type          STRING, #m or d
         default       STRING
       END RECORD
       
DEFINE g_detail_page_link_info   DYNAMIC ARRAY OF RECORD
         table     STRING,
         page      INTEGER,
         var_keys  STRING
       END RECORD
       
DEFINE g_detail_lang_info   DYNAMIC ARRAY OF RECORD
        node              om.DomNode,
        master            STRING,
        master_pk         STRING, 
        master_single_fk  STRING, #160617 - 自身的PK(不含FK)
        master_fk         STRING,
        master_idx        STRING
        END RECORD

DEFINE g_dzbb RECORD 
         prog_name    STRING,
         point_name   STRING,
         point_ver    STRING,
         addpoint     STRING
         END RECORD
         
#+ 單頭多table資料擷取
PUBLIC FUNCTION adzp150_master_multitable_info(p_node)
   DEFINE p_node           om.DomNode
   DEFINE l_child_node     om.DomNode
   DEFINE li_idx           INTEGER
   DEFINE ls_tmp           STRING
   DEFINE ls_tmp2          STRING
   DEFINE ls_tmp3          STRING
   DEFINE li_tmp           INTEGER
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token         STRING
   
   LET g_type = g_properties.getValue("type_id")
   
   #判斷這是第幾個單
   LET li_idx = g_detail_table_info.getLength()+1
   IF li_idx = 1 THEN
      LET li_idx = 2
   END IF
   LET g_master_table_info[li_idx].idx = li_idx
   
   #寫入目前共有幾個單頭table
   CALL g_properties.addAttribute("master_tbl_cnt",li_idx)
   
   #寫入table名稱
   LET ls_tmp = adzp150_create_name(li_idx, "master_tbl_name_by_tbl", "<<<")
   
   CALL g_properties.addAttribute(ls_tmp,p_node.getAttribute("id") )
   
   #開始擷取資料
   
   #table
   LET g_master_table_info[li_idx].id = p_node.getAttribute("id") 
   
   #pk
   LET g_master_table_info[li_idx].pk = p_node.getAttribute("pk")    

   #子節點資料擷取
   LET l_child_node = p_node.getFirstChild()
   WHILE l_child_node IS NOT NULL
      CASE l_child_node.getAttribute("id")
      
         #forupd_sql
         WHEN "forupd_sql"
            LET ls_tmp = adzp150_field_filter(l_child_node.getAttribute("query") ,"sql")
            LET ls_tmp = cl_replace_str(ls_tmp,'FOR UPDATE','')
            LET g_master_table_info[li_idx].forupd_sql = ls_tmp

         #fill_sql
         WHEN "fill_sql"
            LET g_master_table_info[li_idx].fill_sql = adzp150_field_filter(l_child_node.getAttribute("query") ,"sql")

         #多語言table
         WHEN "append"
                  
      END CASE
   
      LET l_child_node = l_child_node.getNext()
      
   END WHILE
   
   #create function
   CALL adzp150_master_create_func(li_idx)
   
   #call select
   LET ls_tmp = g_properties.getValue("master_append_select")
   LET ls_tmp = ls_tmp, 15 SPACES, "CALL ",
                        g_properties.getValue("general_prefix"),"_",g_master_table_info[li_idx].id,"('s')"
   CALL g_properties.addAttribute("master_append_select",ls_tmp)
   
   #call insert
   LET ls_tmp = g_properties.getValue("master_append_insert")
   LET ls_tmp = ls_tmp, 15 SPACES, "CALL ",
                        g_properties.getValue("general_prefix"),"_",g_master_table_info[li_idx].id,"('i')"
   CALL g_properties.addAttribute("master_append_insert",ls_tmp)
   
   #call update
   LET ls_tmp = g_properties.getValue("master_append_update")
   LET ls_tmp = ls_tmp, 15 SPACES, "CALL ",
                        g_properties.getValue("general_prefix"),"_",g_master_table_info[li_idx].id,"('u')"
   CALL g_properties.addAttribute("master_append_update",ls_tmp)
   
   #call delete
   LET ls_tmp = g_properties.getValue("master_append_delete")
   LET ls_tmp = ls_tmp, 15 SPACES, "CALL ",
                        g_properties.getValue("general_prefix"),"_",g_master_table_info[li_idx].id,"('d')"
   CALL g_properties.addAttribute("master_append_delete",ls_tmp)
   
   #call reproduce
   LET ls_tmp = g_properties.getValue("master_append_reproduce")
   FOR li_tmp = 1 TO g_properties.getValue("master_pk_num")
      LET ls_tmp2 = adzp150_create_name(li_tmp, "master_var_pk", "&&")
      LET ls_tmp2 = g_properties.getValue(ls_tmp2)
      LET ls_tmp3 = adzp150_create_name(li_tmp, "master_field_pk", "&&")
      LET ls_tmp3 = g_properties.getValue(ls_tmp3)
      LET ls_tmp = ls_tmp, 3 SPACES, "LET ",ls_tmp2," = l_master.",ls_tmp3,"\n"
      LET ls_tmp = ls_tmp, 3 SPACES, "LET g_",ls_tmp3,"_t = l_master.",ls_tmp3,"\n"
   END FOR
   LET ls_tmp = ls_tmp, 3 SPACES, "CALL ",
                        g_properties.getValue("general_prefix"),"_",g_master_table_info[li_idx].id,"('u')"

   CALL g_properties.addAttribute("master_append_reproduce",ls_tmp)

END FUNCTION     

#+ 單頭多table資料擷取
PRIVATE FUNCTION adzp150_master_create_func(pi_idx)
   DEFINE pi_idx              INTEGER
   DEFINE ls_chk              STRING
   DEFINE ls_select           STRING
   DEFINE ls_insert           STRING
   DEFINE ls_update           STRING
   DEFINE ls_delete           STRING
   DEFINE ls_wc               STRING
   DEFINE lst_token           base.StringTokenizer
   DEFINE ls_token            STRING
   DEFINE ls_fields           STRING
   DEFINE ls_vars             STRING
   DEFINE ls_vars_init        STRING
   DEFINE ls_var_pks          STRING
   DEFINE ls_append_fields    STRING
   DEFINE ls_append_vars      STRING
   DEFINE ls_tmp              STRING
   DEFINE lc_tmp              CHAR(500)
   DEFINE ls_tmp2             STRING
   DEFINE li_tmp              INTEGER 
   DEFINE ls_function         STRING
   DEFINE lc_chk_site_pk      CHAR(500)
   DEFINE ls_chk_site_pk      STRING
   DEFINE ls_key_field        STRING
   
   
   #create append_fields and append_vars
   LET ls_tmp = g_master_table_info[pi_idx].id
   LET ls_tmp2 = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
   LET ls_tmp2 = ls_tmp2, "ent"
   IF cl_getfield(ls_tmp,ls_tmp2) THEN
      LET ls_append_fields = ls_tmp2
      LET ls_append_vars = "g_enterprise"
      LET ls_wc = ls_wc, ls_tmp2, " = g_enterprise AND "
   END IF
   
   #先判斷是否有site
   LET ls_tmp2 = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
   LET ls_tmp2 = ls_tmp2, "site"
   
   #判斷是否為key 
   LET lc_tmp = ls_tmp
   SELECT dzed004 INTO lc_chk_site_pk FROM dzed_t
    WHERE dzed001 = lc_tmp
      AND dzed003 = 'P'
   LET ls_chk_site_pk = lc_chk_site_pk
   
   IF cl_getfield(ls_tmp,ls_tmp2) AND 
      ls_chk_site_pk.getIndexOf(ls_tmp2,1) > 0 THEN
         
      #判斷是否出現在畫面上(general_all_field)
      LET ls_key_field = g_master_table_info[pi_idx].pk
      IF ls_key_field.getIndexOf(ls_tmp2,1) = 0 THEN
         #若無則特別處理
         IF NOT cl_null(ls_append_fields) THEN
            LET ls_append_fields = ls_append_fields, ','
            LET ls_append_vars = ls_append_vars, ','
         END IF
         LET ls_append_fields = ls_append_fields, ls_tmp2
         LET ls_append_vars = ls_append_vars,  "g_site"
         LET ls_wc = ls_wc, ls_tmp2, " = g_site AND "
      END IF
   END IF 
   
   IF NOT cl_null(ls_append_vars) THEN
      LET ls_append_fields = ls_append_fields, ','
      LET ls_append_vars = ls_append_vars, ','
   END IF
   
   #create fields
   LET ls_tmp = g_master_table_info[pi_idx].forupd_sql
   LET ls_fields = ls_tmp.subString( ls_tmp.getIndexOf("SELECT",1)+7, ls_tmp.getIndexOf("FROM",1)-1)
   
   #create vars
   LET lst_token = base.StringTokenizer.create(ls_fields, ',')
   LET ls_tmp = g_properties.getValue("master_var_title"),'.'
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_vars = ls_vars, ls_tmp, ls_token
      LET ls_vars_init = ls_vars_init, 9 SPACES, "LET ",ls_tmp,ls_token, " = NULL \n"
      IF lst_token.hasMoreTokens() THEN
         LET ls_vars = ls_vars, ',' 
      END IF
   END WHILE
   
   #create var pks
   LET ls_var_pks = g_properties.getValue("master_var_allkeys")
      
   #create wc
   LET ls_tmp = g_master_table_info[pi_idx].pk
   LET li_tmp = 1
   LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_tmp = adzp150_create_name(li_tmp, "master_field_pk", "&&") 
      #LET ls_wc = ls_wc, ls_token, ' = g_', g_properties.getValue(ls_tmp), '_t' #170424-00031#1
      LET ls_wc = ls_wc, ls_token, ' = l_', g_properties.getValue(ls_tmp) #170424-00031#1
      IF lst_token.hasMoreTokens() THEN
         LET ls_wc = ls_wc, " AND "
      END IF
      LET li_tmp = li_tmp + 1
   END WHILE
   
   #create chk
   LET ls_chk = "   SELECT COUNT(1) INTO li_cnt FROM ",g_master_table_info[pi_idx].id, " WHERE ", ls_wc
   
   #create select
   LET ls_select = "      LET ls_sql = '",g_master_table_info[pi_idx].forupd_sql,"' \n",
                   "      LET ls_sql = cl_sql_add_mask(ls_sql) \n",
                   "      DECLARE ",g_master_table_info[pi_idx].id,"_cl CURSOR FROM ls_sql \n",
                   "      OPEN ",g_master_table_info[pi_idx].id,"_cl USING ",
                      ls_append_vars,g_properties.getValue("master_var_allkeys"),"\n",
                   "      FETCH ",g_master_table_info[pi_idx].id,"_cl INTO ",ls_vars, '\n'
    
   #create insert
   LET ls_insert = "      INSERT INTO ",g_master_table_info[pi_idx].id," \n",
                   "      (",ls_append_fields,g_master_table_info[pi_idx].pk,",",ls_fields,")\n",
                   "      VALUES (",ls_append_vars,ls_var_pks,",",ls_vars,")"
   
   #create update
   LET ls_update = "      UPDATE ",g_master_table_info[pi_idx].id," SET \n",
                   "      (",g_master_table_info[pi_idx].pk,",",ls_fields,") = \n",
                   "      (",ls_var_pks,",",ls_vars,") \n",
                   "      WHERE ", ls_wc, "\n"

   #create delete
   LET ls_delete = "      DELETE FROM ",g_master_table_info[pi_idx].id,"\n",
                   "      WHERE ", ls_wc
                   
   #組合FUNCTION
   LET g_value = g_master_table_info[pi_idx].id
   CALL g_properties.addAttribute("mdl_tbl_name",g_value)
   LET g_value = ls_select
   CALL g_properties.addAttribute("mdl_select",g_value)
   LET g_value = ls_vars_init
   CALL g_properties.addAttribute("mdl_vars_init",g_value)
   LET g_value = ls_chk
   CALL g_properties.addAttribute("mdl_chk",g_value)
   LET g_value = ls_insert
   CALL g_properties.addAttribute("mdl_insert",g_value)
   LET g_value = ls_update
   CALL g_properties.addAttribute("mdl_update",g_value)
   LET g_value = ls_delete
   CALL g_properties.addAttribute("mdl_delete",g_value)
   LET ls_function = adzp150_make_slice("a06")
    
   #丟進add-point
   LET g_dzbb.prog_name   = g_properties.getValue("app_id")
   LET g_dzbb.point_name  = "function.",g_properties.getValue("general_prefix"),"_",g_master_table_info[pi_idx].id
   LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
   LET g_dzbb.addpoint    = ls_function
   CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     

END FUNCTION

       
#+ 單頭多table資料擷取
PUBLIC FUNCTION adzp150_master_inner_join()
   DEFINE li_idx           INTEGER
   DEFINE ls_tmp           STRING
   DEFINE ls_tmp2          STRING
   DEFINE ls_value         STRING
   DEFINE li_tmp           INTEGER
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token         STRING
   DEFINE lc_dzeb001       LIKE dzeb_t.dzeb001
   DEFINE lc_dzeb002       LIKE dzeb_t.dzeb002
   DEFINE lc_dzeb004       LIKE dzeb_t.dzeb004
      
   FOR li_idx = 2 TO g_master_table_info.getLength()
   
      #判斷單頭table是否有enterprise
      LET ls_tmp = g_master_table_info[li_idx].id
      LET ls_tmp = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      LET ls_tmp = ls_tmp, "ent"
      IF cl_getfield(g_master_table_info[li_idx].id, ls_tmp) THEN
         LET g_master_table_info[li_idx].ent = ls_tmp
      END IF
      
      #判斷單頭table是否有site
      LET ls_tmp = g_master_table_info[li_idx].id
      LET ls_tmp = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      LET ls_tmp = ls_tmp, "site"
      LET ls_tmp2 = g_properties.getValue("general_all_field"),',',
                    g_master_table_info[li_idx].pk
      LET lc_dzeb001 = g_master_table_info[li_idx].id
      LET lc_dzeb002 = ls_tmp
      LET lc_dzeb004 = 'N'
      SELECT dzeb004 INTO lc_dzeb004 FROM dzeb_t 
       WHERE dzeb001 = lc_dzeb001 AND dzeb002 = lc_dzeb002
      IF cl_getfield(g_master_table_info[li_idx].id, ls_tmp) AND
         ls_tmp2.getIndexOf(ls_tmp,1) = 0 AND lc_dzeb004 = "Y" THEN
         LET g_master_table_info[li_idx].site = ls_tmp
      END IF
      
      #join
      LET ls_value = g_properties.getValue("master_join_all")
      LET ls_value = ls_value, " LEFT JOIN ", g_master_table_info[li_idx].id, " ON "
      
      #判斷ent
      IF NOT cl_null(g_master_table_info[li_idx].ent) THEN
         LET ls_value = ls_value, g_master_table_info[li_idx].ent, ' = ',g_properties.getValue("master_field_ent"),' AND '
      END IF
      
      #判斷site
      IF NOT cl_null(g_master_table_info[li_idx].site) THEN
         LET ls_value = ls_value, g_master_table_info[li_idx].site, ' = ',g_properties.getValue("master_field_site"),' AND '
      END IF
      LET lst_token = base.StringTokenizer.create(g_master_table_info[li_idx].pk, ",")
      LET li_tmp = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(li_tmp, "master_field_pk", "&&") 
         LET ls_value = ls_value, g_properties.getValue(ls_tmp), " = ", ls_token
         LET li_tmp = li_tmp + 1
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, " AND "
         END IF
      END WHILE
      CALL g_properties.addAttribute("master_join_all",ls_value)

   END FOR

END FUNCTION
       
#+ 單身多table資料擷取
PUBLIC FUNCTION adzp150_detail_multitable_info(p_node,pi_type)
   DEFINE p_node         om.DomNode
   DEFINE pi_type        INTEGER
   DEFINE l_child_node   om.DomNode
   DEFINE li_idx         INTEGER
   DEFINE li_idx2        INTEGER #160617
   DEFINE ls_tmp         STRING
   DEFINE ls_tmp2        STRING
   DEFINE ls_page        STRING
   DEFINE ls_stus        STRING
   DEFINE lst_token      base.StringTokenizer
   DEFINE ls_token       STRING
   DEFINE ls_value       STRING
   DEFINE ls_name        STRING
   
   LET g_type = g_properties.getValue("type_id")
   
   #判斷是否要處理
   IF pi_type = 1 THEN
      #僅處理第二階單身
      IF NOT cl_null(p_node.getAttribute("master")) THEN
         RETURN
      END IF
   ELSE
      #僅處理第三階單身
      IF cl_null(p_node.getAttribute("master")) THEN
         RETURN
      END IF
   END IF
   
   #判斷這是第幾個單身
   LET li_idx = g_detail_table_info.getLength()+1
   LET g_detail_table_info[li_idx].idx = li_idx
   
   #寫入目前共有幾個單身table
   CALL g_properties.addAttribute("detail_tbl_cnt",li_idx)
   
   #開始擷取資料
   
   #table
   LET g_detail_table_info[li_idx].id = p_node.getAttribute("id") 
   
   #master table #160617
   LET g_detail_table_info[li_idx].master = p_node.getAttribute("master") 
   
   #site_pk
   LET g_detail_table_info[li_idx].site_pk = p_node.getAttribute("default_site") 
   IF cl_null(g_detail_table_info[li_idx].site_pk) THEN
      LET g_detail_table_info[li_idx].site_pk = 'Y'
   END IF 
   
   #因應refernce, 先行取得各table所在page(detail_tbl_name${page})
   LET lst_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_name =  adzp150_create_name(ls_token, "detail_tbl_name", "<<<")
      LET ls_value = p_node.getAttribute("id") 
      CALL g_properties.addAttribute(ls_name, ls_value)
   END WHILE
      
   #page
   IF cl_null(p_node.getAttribute("page")) THEN
      LET g_detail_table_info[li_idx].page = "1"
   ELSE
      LET g_detail_table_info[li_idx].page = p_node.getAttribute("page")
   END IF
   LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_pages", "<<<") 
   CALL g_properties.addAttribute(ls_tmp,g_detail_table_info[li_idx].page)

   #first page
   LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
   LET ls_token = lst_token.nextToken()
   LET g_detail_table_info[li_idx].first_page = ls_token 
   
   #fk
   LET g_detail_table_info[li_idx].fk = p_node.getAttribute("fk") 

   #pk
   LET g_detail_table_info[li_idx].pk = p_node.getAttribute("pk")  
   
   #first_key
   LET lst_token = base.StringTokenizer.create(p_node.getAttribute("pk"), ',')
   LET g_detail_table_info[li_idx].first_key = lst_token.nextToken()

   #allkey
   LET ls_value = g_detail_table_info[li_idx].fk,',',g_detail_table_info[li_idx].pk
   LET ls_name = adzp150_create_name(li_idx, "detail_field_allkeys", "<<<")
   CALL g_properties.addAttribute(ls_name,ls_value) 
   
   #linked
   LET g_detail_table_info[li_idx].linked = p_node.getAttribute("linked")
   
   #add_wc
   LET g_detail_table_info[li_idx].add_wc = p_node.getAttribute("32wc")
   
   #detail_tbl_link
   LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_link", "<<<")
   IF cl_null(g_detail_table_info[li_idx].linked) THEN
      CALL g_properties.addAttribute(ls_tmp,'0')
   ELSE
      CALL g_properties.addAttribute(ls_tmp,g_detail_table_info[li_idx].linked)
   END IF

   #detail(是否為t02的單身)
   LET ls_tmp = p_node.getAttribute("detail")
   
   IF ls_tmp = "Y" AND NOT cl_null(ls_tmp) THEN
      LET ls_tmp = "Y"
      #擷取哪幾個page屬於t02的單身
      LET lst_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_page = adzp150_create_name(ls_token, "general_detail_mark", "<<<") 
         CALL g_properties.addAttribute(ls_page,"#")
         LET ls_page = adzp150_create_name(ls_token, "general_page_detail", "<<<") 
         CALL g_properties.addAttribute(ls_page,"Y")
         LET ls_page = adzp150_create_name(ls_token, "general_page_idx", "<<<") 
         CALL g_properties.addAttribute(ls_page,"2")
         
      END WHILE
   ELSE
      LET ls_tmp = "N"
      #擷取哪幾個page屬於t02的單身
      LET lst_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_page = adzp150_create_name(ls_token, "general_page_detail", "<<<") 
         CALL g_properties.addAttribute(ls_page,"N")
         LET ls_page = adzp150_create_name(ls_token, "general_master_mark", "<<<") 
         CALL g_properties.addAttribute(ls_page,"#")
      END WHILE
   END IF
   LET ls_tmp2 = adzp150_create_name(li_idx, "detail_tbl_type_by_tbl", "<<<")
   CALL g_properties.addAttribute(ls_tmp2,ls_tmp)
   
   #子節點資料擷取
   LET l_child_node = p_node.getFirstChild()
   WHILE l_child_node IS NOT NULL
      CASE l_child_node.getAttribute("id")
      
         #forupd_sql_detail
         WHEN "forupd_sql_detail"
            LET g_detail_table_info[li_idx].forupd_sql = adzp150_field_filter(l_child_node.getAttribute("query") ,"sql")

         #fill_sql
         WHEN "b_fill_sql"
            LET g_detail_table_info[li_idx].fill_sql = adzp150_field_filter(l_child_node.getAttribute("query") ,"sql")

         #多語言table
         WHEN "detail_append"
            LET g_detail_lang_info[g_detail_lang_info.getLength()+1].node           = l_child_node
            LET g_detail_lang_info[g_detail_lang_info.getLength()].master           = g_detail_table_info[li_idx].id
            LET g_detail_lang_info[g_detail_lang_info.getLength()].master_pk        = g_detail_table_info[li_idx].fk,",",g_detail_table_info[li_idx].pk
            #160617
            LET g_detail_lang_info[g_detail_lang_info.getLength()].master_single_fk = g_detail_table_info[li_idx].fk
            IF NOT cl_null(g_detail_table_info[li_idx].master) THEN
               #第三階單身還要額外加上第二階單身的key
               FOR li_idx2 = 1 TO g_detail_table_info.getLength()  
                  IF g_detail_table_info[li_idx2].id = g_detail_table_info[li_idx].master THEN
                     LET g_detail_lang_info[g_detail_lang_info.getLength()].master_fk  = g_detail_table_info[li_idx2].pk,',',g_detail_table_info[li_idx].pk    
                     EXIT FOR
                  END IF
               END FOR
            ELSE            
               LET g_detail_lang_info[g_detail_lang_info.getLength()].master_fk  = g_detail_table_info[li_idx].pk    
            END IF
            LET g_detail_lang_info[g_detail_lang_info.getLength()].master_idx = li_idx
            #160617
            #CALL adzp150_detail_multi_table(l_child_node)
            #CALL adzp150_update_item_action('d',l_child_node)
                  
      END CASE
   
      LET l_child_node = l_child_node.getNext()
   END WHILE
   
   #確定該單身的類型
   IF p_node.getAttribute("detail") = "Y" AND NOT cl_null(p_node.getAttribute("detail")) THEN
      CALL adzp150_detail2_info(p_node,li_idx)
   ELSE
      CALL adzp150_detail_info(p_node,li_idx)
   END IF

END FUNCTION

#+ 單身多table入口
PUBLIC FUNCTION adzp150_create_detail_entrance()

   CALL g_properties.addAttribute("detail_table_num",g_detail_table_info.getLength())

   #確定table間的分組關係
   CALL adzp150_create_table_group()

   #先將table基本資料轉為table為主的資料
   CALL adzp150_create_advanced_table_info()
   
   #取得獨立的page資料
   CALL adzp150_create_independent_page_info()
   
   #再將table為主的資料轉為page為主的資料
   CALL adzp150_create_advanced_page_info()
   
   #多語言處理
   CALL adzp150_create_lang()
   
   #key欄位chk段落
   CALL adzp150_create_chk()
   
   #單身link特殊處理
   CALL adzp150_create_page_link_info()
   
   #detail_vars_all重製
   CALL adzp150_create_detail_vars_all()

END FUNCTION

#+ 確定table間的分組關係
PRIVATE FUNCTION adzp150_create_table_group()
   DEFINE ls_id        STRING   
   DEFINE ls_page      STRING
   DEFINE ls_link      STRING
   DEFINE ls_tmp       STRING
   DEFINE ls_name      STRING
   DEFINE ls_value     STRING
   DEFINE li_idx       INTEGER
   DEFINE li_cnt       INTEGER
   DEFINE li_idx2      INTEGER
   DEFINE li_cnt2      INTEGER
   DEFINE lb_chk       BOOLEAN
   DEFINE lst_token    base.StringTokenizer
   DEFINE ls_token     STRING
   
   LET li_cnt = g_detail_table_info.getLength()
   
   FOR li_idx = 1 TO li_cnt
   
      LET ls_id   = g_detail_table_info[li_idx].id
      LET ls_page = g_detail_table_info[li_idx].page
      LET ls_link = g_detail_table_info[li_idx].page
      
      #根據link分組
      LET li_cnt2 = g_table_group.getLength()
      IF li_cnt2 = 0 THEN
         #尚無資料時訂為第一組
         LET g_table_group[li_cnt2+1].id = ls_id,","
         LET g_table_group[li_cnt2+1].page = ls_page,","
      ELSE
         LET lb_chk = TRUE
         FOR li_idx2 = 1 TO li_cnt2
            LET ls_tmp = g_table_group[li_idx2].page
            IF ls_tmp.getIndexOf(g_detail_table_info[li_idx].linked,1) > 0 THEN
               LET g_table_group[li_idx2].id = g_table_group[li_idx2].id,ls_id,","
               LET g_table_group[li_idx2].page = g_table_group[li_idx2].page,ls_page,","
               LET lb_chk = FALSE
            END IF
         END FOR
         IF lb_chk THEN
            LET g_table_group[li_cnt2+1].id = ls_id,","
            LET g_table_group[li_cnt2+1].page = ls_page,","
         END IF
      END IF
      
   END FOR
   
   #整理page分組
   FOR li_idx = 1 TO g_table_group.getLength()
      LET ls_tmp = ""
      LET lst_token = base.StringTokenizer.create(g_table_group[li_idx].page, ",")
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = ls_tmp, "'", ls_token, "',"
      END WHILE
      LET g_table_group[li_idx].page2 = ls_tmp
   END FOR
   
   #同分組的資料刪除時連帶刪除其他頁面
   FOR li_idx = 1 TO g_table_group.getLength()
      LET ls_tmp = ""
      LET lst_token = base.StringTokenizer.create(g_table_group[li_idx].page, ",")
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_var_title", "<<<")
         LET ls_tmp = ls_tmp, "IF ps_page <> \"'",ls_token,"'\" THEN \n      ",
                              "   CALL ", g_properties.getValue(ls_name), ".deleteElement(li_idx) \n      ",
                              "END IF \n      "
      END WHILE
      LET g_table_group[li_idx].page_delete = ls_tmp
   END FOR
   
   #同分組的資料新增時連帶新增其他頁面
   FOR li_idx = 1 TO g_table_group.getLength()
      LET ls_tmp = ""
      LET lst_token = base.StringTokenizer.create(g_table_group[li_idx].page, ",")
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_var_title", "<<<")
         LET ls_tmp = ls_tmp, "IF ps_page <> \"'",ls_token,"'\" THEN \n      ",
                              "   CALL ", g_properties.getValue(ls_name), ".insertElement(li_idx) \n      ",
                              "END IF \n      "
      END WHILE
      LET g_table_group[li_idx].page_insert = ls_tmp
   END FOR
   
   #同分組的資料修改key時連帶修改其他頁面
   FOR li_idx = 1 TO g_table_group.getLength()
      LET ls_tmp = ""
      LET lst_token = base.StringTokenizer.create(g_table_group[li_idx].page, ",")
      #將不同page的key資料進行覆蓋
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_var_title", "<<<")
         LET ls_tmp = ls_tmp, "IF ps_page <> \"'",ls_token,"'\" THEN \n      ",
                              "   CALL ", g_properties.getValue(ls_name), ".insertElement(li_idx) \n      ",
                              "END IF \n      "
      END WHILE
      LET g_table_group[li_idx].page_insert = ls_tmp
   END FOR
   
   LET li_cnt = g_detail_table_info.getLength()
   
   #單身table分組
   FOR li_idx = 1 TO li_cnt
      IF g_detail_table_info[li_idx].type = "D" THEN
         FOR li_idx2 = 1 TO g_table_group.getLength()
            LET ls_tmp = g_table_group[li_idx2].id
            IF ls_tmp.getIndexOf(g_detail_table_info[li_idx].master,1) THEN
               LET ls_name = adzp150_create_name(li_idx, "detail_tbl_group", "<<<")
               LET ls_value = g_table_group[li_idx2].page2
               CALL g_properties.addAttribute(ls_name, ls_value)
            END IF
         END FOR
      END IF
   END FOR

END FUNCTION

#+ 先將table基本資料轉為table為主的資料
PRIVATE FUNCTION adzp150_create_advanced_table_info()
   DEFINE li_cnt          INTEGER
   DEFINE li_idx          INTEGER
   DEFINE li_idx2         INTEGER
   DEFINE li_idx3         INTEGER
   DEFINE li_tmp          INTEGER
   DEFINE ls_tmp          STRING
   DEFINE lc_tmp          CHAR(500)
   DEFINE lc_tmp2         CHAR(500)
   DEFINE ls_tmp2         STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE lst_token2      base.StringTokenizer
   DEFINE ls_token2       STRING
   DEFINE lb_chk          BOOLEAN
   DEFINE ls_site_chk     STRING
   DEFINE lb_site_chk     BOOLEAN
   DEFINE ls_chk_site_pk  STRING
   DEFINE lc_chk_site_pk  CHAR(500)
   DEFINE li_tbl_idx      INTEGER
   DEFINE li_tbl_idx2     INTEGER
   DEFINE l_tok           base.StringTokenizer #250527-00018#1 add    
   DEFINE l_str           STRING               #250527-00018#1 add
   
   LET li_cnt = g_detail_table_info.getLength()
   
   #進行資料轉換
   FOR li_idx = 1 TO li_cnt
   
      #先判斷table是否重複
      LET lb_chk = TRUE
      IF li_idx > 1 THEN
         FOR li_idx3 = 1 TO li_idx - 1
            IF g_detail_table_info[li_idx].id = g_detail_table_info[li_idx3].id THEN
               LET lb_chk = FALSE
               LET ls_name = adzp150_create_name(li_idx, "general_mark_tbl", "<<<")
               LET ls_value = "#"
               CALL g_properties.addAttribute(ls_name,ls_value)
            END IF
         END FOR
      END IF
   
      #判斷單身table是否有enterprise
      LET ls_tmp = g_detail_table_info[li_idx].id
      LET ls_tmp = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      LET ls_tmp = ls_tmp, "ent"
      IF cl_getfield(g_detail_table_info[li_idx].id, ls_tmp) THEN
         LET g_detail_table_info[li_idx].ent = ls_tmp
         LET ls_name = adzp150_create_name(li_idx, "detail_field_ent", "<<<")
         LET ls_value = ls_tmp
         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF
      
      #判斷單身table是否有site
      LET ls_tmp = g_detail_table_info[li_idx].id
      LET ls_tmp = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      LET ls_tmp = ls_tmp, "site"
      LET ls_site_chk = g_detail_table_info[li_idx].pk,',',g_detail_table_info[li_idx].fk
      LET ls_tmp2 = g_properties.getValue("general_all_field")
      IF ls_site_chk.getIndexOf(ls_tmp,1) = 0 AND
         ls_tmp2.getIndexOf(ls_tmp,1) = 0 THEN
         LET lb_site_chk = TRUE
      ELSE
         LET lb_site_chk = FALSE
      END IF
      
      #判定單身site是否為pk
      LET lc_tmp = g_detail_table_info[li_idx].id
      SELECT dzed004 INTO lc_chk_site_pk FROM dzed_t
       WHERE dzed001 = lc_tmp
         AND dzed003 = 'P'
      
      LET ls_chk_site_pk = lc_chk_site_pk
      IF cl_getfield(g_detail_table_info[li_idx].id, ls_tmp) AND
         ls_chk_site_pk.getIndexOf(ls_tmp,1) > 0 AND
         lb_site_chk THEN
         LET g_detail_table_info[li_idx].site = ls_tmp
         LET ls_name = adzp150_create_name(li_idx, "detail_field_site", "<<<")
         LET ls_value = ls_tmp
         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF
   
      #detail_pk_num
      LET li_tmp = 0
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].pk, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET li_tmp = li_tmp + 1
      END WHILE  
      LET ls_value = li_tmp USING "<<<"
      LET ls_name = adzp150_create_name(li_idx, "detail_pk_num_by_tbl", "<<<") 
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      IF li_idx = 1 AND (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
         LET ls_name = "master_pk_num"
         CALL g_properties.addAttribute("master_pk_num", ls_value)
      END IF
   
      #detail_fields_default
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(ls_token, "detail_fields_default", "<<<") 
         IF NOT cl_null(g_properties.getValue(ls_tmp)) THEN
            LET ls_value = ls_value, g_properties.getValue(ls_tmp), "\n"
         END IF
      END WHILE     
      
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET g_detail_page_info[ls_token].default = ls_value
      END WHILE     
    
      #detail_fill_sql
      LET ls_value = g_detail_table_info[li_idx].fill_sql
      LET ls_name = adzp150_create_name(li_idx, "detail_fill_sql", "<<<") 
      LET ls_value = adzp150_detail_sql_forupd(g_detail_table_info[li_idx].fill_sql,g_detail_table_info[li_idx].page,g_detail_table_info[li_idx].id)
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #detail_fill_sql_pre
      LET ls_tmp = ls_value
      LET ls_value = ls_tmp.subString(1,ls_tmp.getIndexOf("FROM",1)-1)
      LET ls_value = ls_value.subString(1,ls_tmp.getIndexOf("SELECT",1)+6),
                     " DISTINCT ",
                     adzp150_add_table_name(ls_value.subString(ls_tmp.getIndexOf("SELECT",1)+7,ls_value.getLength()),'d')
      LET ls_name = adzp150_create_name(li_idx, "detail_fill_sql_pre", "<<<") 
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_fill_sql_post
      LET ls_value = ls_tmp.subString(ls_tmp.getIndexOf("WHERE",1)-1,ls_tmp.getLength())
      LET ls_value = adzp150_add_table_name_wc(ls_value,'d')
      LET ls_name = adzp150_create_name(li_idx, "detail_fill_sql_post", "<<<") 
      CALL g_properties.addAttribute(ls_name,ls_value)

      #detail_sql_forupd
      LET ls_value = g_detail_table_info[li_idx].forupd_sql
      LET ls_name = adzp150_create_name(li_idx, "detail_sql_forupd", "<<<") 
      LET ls_value = adzp150_detail_sql_forupd(g_detail_table_info[li_idx].forupd_sql,g_detail_table_info[li_idx].page,g_detail_table_info[li_idx].id)
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #組合${detail_table_bcl${table}}
      LET ls_tmp = g_properties.getValue("general_prefix"), "_bcl"
      LET ls_value = adzp150_create_name(li_idx, ls_tmp, "<<<") 
      LET ls_name = adzp150_create_name(li_idx, "detail_table_bcl", "<<<") 
      CALL g_properties.addAttribute(ls_name, ls_value)
   
      #取得table名稱(by table)
      LET ls_value = g_detail_table_info[li_idx].id
      LET ls_name = adzp150_create_name(li_idx, "detail_tbl_name_by_tbl", "<<<") 
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #190730-00003#4 -S-
      LET ls_value = ls_value.subString(1,ls_value.getIndexOf('_',1)-1)
      LET ls_name = adzp150_create_name(li_idx, "detail_tbl_pre_name_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name, ls_value)
      #190730-00003#4 -E-
      
      #fk欄位相關
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].fk, ',')
      LET li_idx2 = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #取得各別fk欄位名稱
         LET ls_name = adzp150_create_name(li_idx, "detail_by_tbl", "<<<")
         LET ls_name = adzp150_create_name(li_idx2, ls_name||"_field_fk", "&&")
         LET ls_value = ls_token
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #取得各別fk變數名稱
         LET ls_name = adzp150_create_name(li_idx, "detail_by_tbl", "<<<")
         LET ls_name = adzp150_create_name(li_idx2, ls_name||"_var_fk", "&&")
         LET ls_tmp = adzp150_create_name(li_idx2, "master_field_pk", "&&")
         LET ls_value = g_properties.getValue(ls_tmp)
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         LET li_idx2 = li_idx2 + 1
      END WHILE
      
      #pk欄位相關
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].pk, ',')
      LET li_idx2 = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #取得各別fk欄位名稱
         LET ls_name = adzp150_create_name(li_idx, "detail_by_tbl", "<<<")
         LET ls_name = adzp150_create_name(li_idx2, ls_name||"_field_pk", "&&")
         LET ls_value = ls_token
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         LET li_idx2 = li_idx2 + 1
      END WHILE
      
      #若無重複則組合 left join
      IF lb_chk THEN
         LET ls_value = " LEFT JOIN ", g_detail_table_info[li_idx].id, " ON "
         
         #判斷ent
         IF NOT cl_null(g_detail_table_info[li_idx].ent) THEN
            #t02特別處理
            IF (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
               LET ls_value = ls_value, g_detail_table_info[li_idx].ent, ' = ', g_properties.getValue("detail_field_ent"),' AND '
            ELSE
               LET ls_value = ls_value, g_detail_table_info[li_idx].ent, ' = ', g_properties.getValue("master_field_ent"),' AND '
            END IF
         END IF
         
         #判斷site
         IF NOT cl_null(g_detail_table_info[li_idx].site) THEN
            IF (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
               LET ls_value = ls_value, g_detail_table_info[li_idx].site, ' = ', g_properties.getValue("detail_field_site"),' AND '
            ELSE
               LET ls_value = ls_value, g_detail_table_info[li_idx].site, ' = ', g_properties.getValue("master_field_site"),' AND '
            END IF
         END IF
         
         IF cl_null(g_detail_table_info[li_idx].fk) THEN
            LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].pk, ",")
         ELSE
            LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].fk, ",")
         END IF
         
         LET lst_token2 = base.StringTokenizer.create(g_detail_table_info[li_idx].mapping_fk, ",")
         
         LET li_tmp = 1
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_token2 = lst_token2.nextToken()
            LET ls_value = ls_value, ls_token2, " = ", ls_token
            LET li_tmp = li_tmp + 1
            IF lst_token.hasMoreTokens() THEN
               LET ls_value = ls_value, " AND "
            END IF
         END WHILE
         LET ls_name = adzp150_create_name(li_idx, "detail_join_by_table", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)
      END IF
       
      #detail_append_wc_s_by_tbl
      LET ls_value = ""
      IF NOT cl_null(g_detail_table_info[li_idx].ent) THEN
         LET ls_value = ls_value, g_detail_table_info[li_idx].ent , " = \" ||g_enterprise|| \" AND "
      END IF
      IF NOT cl_null(g_detail_table_info[li_idx].site) THEN
         LET ls_value = ls_value, g_detail_table_info[li_idx].site , " = '\" ||g_site|| \"' AND "
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_append_wc_s_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_append_wc_by_tbl
      LET ls_value = ""
      IF NOT cl_null(g_detail_table_info[li_idx].ent) THEN
         LET ls_value = ls_value, g_detail_table_info[li_idx].ent, " = g_enterprise AND "
      END IF
      IF NOT cl_null(g_detail_table_info[li_idx].site) THEN
         LET ls_value = ls_value, g_detail_table_info[li_idx].site, " = g_site AND "
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_append_wc_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_field_append_by_tbl
      LET ls_value = ""
      IF NOT cl_null(g_detail_table_info[li_idx].ent) THEN
         LET ls_value = ls_value, g_detail_table_info[li_idx].ent, ", "
      END IF
      IF NOT cl_null(g_detail_table_info[li_idx].site) THEN
         LET ls_value = ls_value, g_detail_table_info[li_idx].site, ", "
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_field_append_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_var_append_by_tbl
      LET ls_value = ""
      IF NOT cl_null(g_detail_table_info[li_idx].ent) THEN
         LET ls_value = ls_value, "g_enterprise, "
      END IF
      IF NOT cl_null(g_detail_table_info[li_idx].site) THEN
         LET ls_value = ls_value, "g_site, "
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_var_append_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #create detail_name_by_table
      LET ls_name = adzp150_create_name(li_idx, "detail_name_by_table", "<<<")
      LET ls_value = g_detail_table_info[li_idx].id
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #create detail_join_name_by_table
      LET ls_name = adzp150_create_name(li_idx, "detail_join_name_by_table", "<<<")
      LET li_tbl_idx2 = 1
      FOR li_tbl_idx = 1 TO li_idx - 1
         #代表定義過相同table
         IF g_detail_table_info[li_idx].id = g_detail_table_info[li_tbl_idx].id THEN
            LET li_tbl_idx2 = li_tbl_idx2 + 1
         END IF
      END FOR
      LET ls_value = g_detail_table_info[li_idx].id, li_tbl_idx2 USING "<<<"
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #detail_32wc_by_table
      LET ls_name = adzp150_create_name(li_idx, "detail_32wc_by_table", "<<<")
      LET ls_value = g_detail_table_info[li_idx].add_wc
      #若有32wc需在前頭加上AND
      IF NOT cl_null(ls_value) THEN
         LET ls_value = ' AND ', ls_value
      END IF
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #create detail_table_group
      LET li_tmp = 1
      LET ls_value = ""
      FOR li_tmp = 1 TO g_table_group.getLength()
         LET ls_value = g_table_group[li_tmp].id         
         IF ls_value.getIndexOf(g_detail_table_info[li_idx].id,1) > 0 THEN
            LET ls_name = adzp150_create_name(li_idx, "detail_table_group", "<<<")
            CALL g_properties.addAttribute(ls_name, ls_value)
            EXIT FOR
         END IF
      END FOR 
      
      #create detail_page_group
      LET li_tmp = 1
      LET ls_value = ""
      FOR li_tmp = 1 TO g_table_group.getLength()
         LET ls_value = g_table_group[li_tmp].page         
         IF ls_value.getIndexOf(g_detail_table_info[li_idx].page,1) > 0 THEN
            LET ls_name = adzp150_create_name(li_idx, "detail_page_group", "<<<")
            LET ls_value = g_table_group[li_tmp].page2
            CALL g_properties.addAttribute(ls_name, ls_value)
            EXIT FOR
         END IF
      END FOR 
      
      #create detail_page_group_delete
      LET li_tmp = 1
      LET ls_value = ""
      FOR li_tmp = 1 TO g_table_group.getLength()
         LET ls_value = g_table_group[li_tmp].page         
         IF ls_value.getIndexOf(g_detail_table_info[li_idx].page,1) > 0 THEN
            LET ls_name = adzp150_create_name(li_idx, "detail_page_group_delete", "<<<")
            LET ls_value = g_table_group[li_tmp].page_delete
            CALL g_properties.addAttribute(ls_name, ls_value)
            EXIT FOR
         END IF
      END FOR 
      
      #create detail_page_group_insert
      LET li_tmp = 1
      LET ls_value = ""
      FOR li_tmp = 1 TO g_table_group.getLength()
         LET ls_value = g_table_group[li_tmp].page         
         IF ls_value.getIndexOf(g_detail_table_info[li_idx].page,1) > 0 THEN
            LET ls_name = adzp150_create_name(li_idx, "detail_page_group_insert", "<<<")
            LET ls_value = g_table_group[li_tmp].page_insert
            CALL g_properties.addAttribute(ls_name, ls_value)
            EXIT FOR
         END IF
      END FOR 
      
      #detail_tbl_fields_qbe
      IF g_type.subString(1,1) <> "q" THEN
         LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page,",")
         LET ls_value = ""
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_name = "detail_fields_qbe",ls_token
            LET ls_tmp = g_properties.getValue(ls_name)
            LET ls_value = ls_value, ls_tmp
            IF lst_token.hasMoreTokens() THEN
               LET ls_value = ls_value, ","
            END IF
         END WHILE
         LET ls_value = adzp150_detail_Filter(ls_value,FALSE)
         IF cl_null(ls_value) THEN
            LET ls_value = 'NULL'
         END IF
         LET ls_name = adzp150_create_name(li_idx, "detail_tbl_fields_qbe", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)
      END IF
      
      #detail_tbl_srfield_all
      IF g_type.subString(1,1) <> "q" THEN
         LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page,",")
         LET ls_value = ""
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_name = "detail_srfield_all", ls_token
            LET ls_tmp = g_properties.getValue(ls_name)
            LET ls_value = ls_value, ls_tmp
            IF lst_token.hasMoreTokens() THEN
               LET ls_value = ls_value, ","
            END IF
         END WHILE

         LET lst_token = base.StringTokenizer.create(ls_value,",")
         LET ls_value = ""
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            #230107-00001#1前方打點,避免l_abcd001與abcd001 共存時失效
            LET ls_tmp = ".",ls_token.subString(ls_token.getIndexOf(".",1)+1,ls_token.getLength()),','
            IF ls_value.getIndexOf(ls_tmp,1) = 0 THEN
               LET ls_value = ls_value, ls_token, ","
            ELSE #DISPLAY 'INFO ->不合規矩:',ls_value,' 字串裡面有',ls_tmp
            END IF
         END WHILE
         LET ls_value = ls_value.subString(1,ls_value.getLength()-1)
         LET ls_name = adzp150_create_name(li_idx, "detail_tbl_srfield_all", "<<<")
         IF cl_null(ls_value) THEN
            LET ls_value = 'NULL'
         END IF
         CALL g_properties.addAttribute(ls_name, ls_value)
      END IF
      
      #detail_tbl_fields_ctrlp
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page,",")
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = "detail_fields_ctrlp", ls_token
         LET ls_tmp = g_properties.getValue(ls_name)
         LET ls_value = ls_value, ls_tmp
      END WHILE
      LET ls_value = adzp150_detail_Filter(ls_value,TRUE)
      LET ls_name = adzp150_create_name(li_idx, "detail_tbl_fields_ctrlp", "<<<")
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      
      #create detail_field_fks
      LET ls_name = adzp150_create_name(li_idx, "detail_field_fks_by_tbl", "<<<")
      LET ls_value = g_detail_table_info[li_idx].fk
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #create detail_field_pks
      LET ls_name = adzp150_create_name(li_idx, "detail_field_pks_by_tbl", "<<<")
      LET ls_value = g_detail_table_info[li_idx].pk
      
      LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_type_by_tbl", "<<<")
      LET ls_tmp = g_properties.getValue(ls_tmp)
      IF cl_null(ls_tmp) THEN
         LET ls_tmp = "N"
      END IF
      IF (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") AND ls_tmp = "Y" THEN
         LET ls_value = g_detail_table_info[li_idx].fk, ',', ls_value
      END IF
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #create detail_field_pks
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].pk,",")
      LET li_idx2 = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(li_idx, "detail", "<<<")
         LET ls_name = ls_name, "_field_pk"
         LET ls_name = adzp150_create_name(li_idx2, ls_name, "&&")
         LET ls_name = ls_name, "_by_tbl"
         LET ls_value = ls_token
         CALL g_properties.addAttribute(ls_name, ls_value)
         LET li_idx2 = li_idx2 + 1
      END WHILE

      #create detail_field_others
      LET ls_value = ""
      LET ls_name = adzp150_create_name(li_idx, "detail_field_others", "<<<")
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = "detail_page_field", ls_token
         LET ls_value = ls_value, g_properties.getValue(ls_tmp), ","
      END WHILE
      LET ls_tmp = g_detail_table_info[li_idx].id
      LET ls_tmp = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      LET ls_tmp2 = g_detail_table_info[li_idx].fk,",",g_detail_table_info[li_idx].pk
      LET lst_token = base.StringTokenizer.create(ls_value, ',')
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         IF ls_token.getIndexOf("(",1) THEN
            LET ls_token = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
         END IF
         IF adzp150_chk_intbl(g_detail_table_info[li_idx].id,ls_token) AND 
            ls_tmp2.getIndexOf(ls_token,1) = 0 THEN
            LET ls_value = ls_value, ls_token, ","
         END IF
      END WHILE
      IF NOT cl_null(ls_value) THEN
         LET ls_value = ",",ls_value.subString(1,ls_value.getLength()-1)
      END IF
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #create detail_var_others
      LET ls_name = adzp150_create_name(li_idx, "detail_field_others", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET lst_token = base.StringTokenizer.create(ls_value, ',')
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         FOR li_tmp = 1 TO g_properties.getValue("page")
           #增加一層判斷, 若不是該table所在的頁面則跳過
           LET ls_tmp = g_detail_table_info[li_idx].page
           IF NOT ls_tmp.getIndexOf(li_tmp,1) THEN
              CONTINUE FOR
           END IF
           LET ls_tmp = "detail_page_field", li_tmp USING "<<<"
           LET ls_tmp = g_properties.getValue(ls_tmp)
           IF ls_tmp.getIndexOf(ls_token,1) > 0 THEN
              #250527-00018#1 add --- start ---
              # 增加一層判斷，避免部分符合就套用狀況
              LET l_tok = base.StringTokenizer.createExt(ls_tmp,",","",TRUE)
              WHILE l_tok.hasMoreTokens()	#依序取得子字串
                 LET l_str = l_tok.nextToken()
                 IF l_str.getIndexOf("(", 1) > 1 THEN
                    LET l_str = l_str.subString(1, l_str.getIndexOf("(", 1) -1)
                 END IF
                 IF l_str = ls_token THEN
                    LET ls_tmp = adzp150_create_name(li_tmp, "detail_var_title", "<<<")
                    EXIT FOR
                 END IF
              END WHILE  
              #250527-00018#1 add --- end ---
              #LET ls_tmp = adzp150_create_name(li_tmp, "detail_var_title", "<<<") #250527-00018#1 mark
              #EXIT FOR #250527-00018#1 mark
           END IF
         END FOR
         LET ls_value = ls_value, g_properties.getValue(ls_tmp), g_detail_table_info[li_idx].curr_idx, ls_token, ","
      END WHILE
      IF NOT cl_null(ls_value) THEN
         LET ls_value = ",",ls_value.subString(1,ls_value.getLength()-1)
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_var_others", "<<<")
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #detail_by_tbl_var_pks
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(g_detail_page_info[g_detail_table_info[li_idx].first_page].var_pk, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ","
         END IF
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_by_tbl_var_pks", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)

      #detail_var_fks
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].var_fk, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ","
         END IF
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_var_fks_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_del_wc_by_tbl
      LET ls_value = ""
      LET ls_tmp = g_detail_table_info[li_idx].fk
      LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
      LET ls_tmp = g_properties.getValue(adzp150_create_name(li_idx, "detail_var_fks_by_tbl", "<<<"))        
      LET lst_token2 = base.StringTokenizer.create(ls_tmp, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_token2 = lst_token2.nextToken()
         LET ls_value = ls_value, ls_token, " = ", ls_token2
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, " AND "
         END IF
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_del_wc_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_var_ps_keys
      LET ls_tmp = g_detail_table_info[li_idx].fk,",",g_detail_table_info[li_idx].pk
      LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
      LET ls_value = ""
      LET li_tmp = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         &ifdef is_ora
         LET ls_value = ls_value, "ps_keys[",li_tmp USING "<<<","]"
         &endif
         &ifdef is_pgs
         #PGS-02667 (200804-00006)  -S-
         IF adzp150_chk_int_filed(ls_token) THEN
            #postgreSQL如果是數字key需額外轉型
            LET ls_value = ls_value, "CAST(ps_keys[",li_tmp USING "<<<","] as int) "
         ELSE
            LET ls_value = ls_value, "ps_keys[",li_tmp USING "<<<","]"
         END IF
         #PGS-02667 (200804-00006)  -E-
         &endif
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ","
         END IF
         LET li_tmp = li_tmp + 1
      END WHILE 
      LET ls_name = adzp150_create_name(li_idx, "detail_var_ps_keys", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_var_fs_keys
      LET ls_tmp = g_detail_table_info[li_idx].fk
      LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
      LET ls_value = ""
      LET li_tmp = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         &ifdef is_ora
         LET ls_value = ls_value, "ps_keys[",li_tmp USING "<<<","]"
         &endif
         &ifdef is_pgs
         #PGS-02667 (200804-00006)  -S-
         IF adzp150_chk_int_filed(ls_token) THEN
            #postgreSQL如果是數字key需額外轉型
            LET ls_value = ls_value, "CAST(ps_keys[",li_tmp USING "<<<","] as int) "
         ELSE
            LET ls_value = ls_value, "ps_keys[",li_tmp USING "<<<","]"
         END IF
         #PGS-02667 (200804-00006)  -E-
         &endif
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ","
         END IF
         LET li_tmp = li_tmp + 1
      END WHILE 
      LET ls_name = adzp150_create_name(li_idx, "detail_var_ps_fkeys", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
         
      #detail_tbl_wc
      LET ls_tmp = g_detail_table_info[li_idx].fk,",",g_detail_table_info[li_idx].pk
      LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
      LET ls_value = ""
      LET li_tmp = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #PGS-02667 (200804-00006)  -S-
         &ifdef is_ora
         LET ls_value = ls_value, ls_token, " = ps_keys_bak[",li_tmp USING "<<<","]"
         &endif
         &ifdef is_pgs
         IF adzp150_chk_int_filed(ls_token) THEN
            #postgreSQL如果是數字key需額外轉型
            LET ls_value = ls_value, ls_token, " = CAST(ps_keys_bak[",li_tmp USING "<<<","] as int) " 
         ELSE
            LET ls_value = ls_value, ls_token, " = ps_keys_bak[",li_tmp USING "<<<","]" 
         END IF
         &endif
         #PGS-02667 (200804-00006)  -E-
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, " AND "
         END IF
         LET li_tmp = li_tmp + 1
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_tbl_wc", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_tbl_wc_by_body
      LET ls_tmp = g_detail_table_info[li_idx].fk
      LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
      LET ls_value = ""
      LET li_tmp = 1
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #PGS-02667 (200804-00006)  -S-
         &ifdef is_ora
         LET ls_value = ls_value, ls_token, " = ps_keys_bak[",li_tmp USING "<<<","]"
         &endif
         &ifdef is_pgs
         IF adzp150_chk_int_filed(ls_token) THEN
            #postgreSQL如果是數字key需額外轉型
            LET ls_value = ls_value, ls_token, " = CAST(ps_keys_bak[",li_tmp USING "<<<","] as int) " 
         ELSE
            LET ls_value = ls_value, ls_token, " = ps_keys_bak[",li_tmp USING "<<<","]" 
         END IF
         #PGS-02667 (200804-00006)  -E-
         &endif
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, " AND "
         END IF
         LET li_tmp = li_tmp + 1
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_tbl_wc_by_body", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_vars_all_by_tbl
      CALL adzp150_create_detail_vars_all_pre()
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(ls_token, "detail_vars_all", "<<<") 
         IF NOT cl_null(g_properties.getValue(ls_tmp)) THEN
            LET ls_value = ls_value, g_properties.getValue(ls_tmp), ","
         END IF
      END WHILE
      LET ls_value = ls_value.subString(1,ls_value.getLength()-1)
      LET ls_name = adzp150_create_name(li_idx, "detail_vars_all_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_field_order_by_tbl
      LET ls_value = ""
      LET ls_tmp = g_detail_table_info[li_idx].pk
      LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
      
      IF (
         g_properties.getValue("type_id") = "i04" OR 
         g_properties.getValue("type_id") = "i07" OR 
         g_properties.getValue("type_id") = "i08" OR 
         g_properties.getValue("type_id") = "t01"  
         ) THEN
         &ifdef is_ora
         LET ls_tmp = g_detail_table_info[li_idx].id #PGS-02667 (200804-00006) #mark
         &endif
         &ifdef is_pgs
         LET ls_tmp = "" #PGS-02667 (200804-00006) #add
         &endif
      ELSE
         &ifdef is_ora
         LET ls_tmp = 't0' #PGS-02667 (200804-00006) #mark
         &endif
         &ifdef is_pgs
         LET ls_tmp = 't0.' #PGS-02667 (200804-00006) #add
         &endif
      END IF
      
      LET ls_value = ''
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         &ifdef is_ora
         LET ls_value = ls_value, ls_tmp, '.', ls_token #PGS-02667 (200804-00006) #mark
         &endif
         &ifdef is_pgs
         LET ls_value = ls_value, ls_tmp, ls_token #PGS-02667 (200804-00006) #add
         &endif
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ','
         END IF
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_field_order_by_tbl", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #Q類資料配置
      IF g_type.subString(1,1) = 'q' THEN
         CALL adzp150_create_query_entrance(li_idx)
      END IF
      
   END FOR

END FUNCTION

#+ 取得獨立的page資料
PRIVATE FUNCTION adzp150_create_independent_page_info()
   DEFINE li_cnt       INTEGER
   DEFINE li_idx       INTEGER
   DEFINE li_idx2      INTEGER
   DEFINE li_tmp       INTEGER
   DEFINE ls_tmp       STRING
   DEFINE ls_tmp2      STRING
   DEFINE ls_name      STRING
   DEFINE ls_value     STRING
   DEFINE li_page      INTEGER
   DEFINE lst_token    base.StringTokenizer
   DEFINE ls_token     STRING
   DEFINE lst_token2   base.StringTokenizer
   DEFINE ls_token2    STRING
   DEFINE ls_first_key STRING
    
   LET li_cnt = g_detail_table_info.getLength()
   FOR li_idx = 1 TO li_cnt
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      LET li_tmp = 1
      WHILE lst_token.hasMoreTokens()
         LET li_page = lst_token.nextToken()
         #單身table, 不同page特別處理
         IF li_tmp > 1 THEN
            LET ls_first_key = g_detail_table_info[li_idx].pk
            IF ls_first_key.getIndexOf(',',1) > 0 THEN
               LET ls_first_key = ls_first_key.subString(1,ls_first_key.getIndexOf(',',1)-1)
            END IF
            LET ls_value = "LET g_insert = 'Y' \n",
                           "            NEXT FIELD ",ls_first_key,"\n"
            LET ls_name = adzp150_create_name(li_page, "detail_page_insert", "<<<")
            CALL g_properties.addAttribute(ls_name,ls_value)
            LET ls_value = "FALSE"
            LET ls_name = adzp150_create_name(li_page, "detail_allow_insert", "<<<")
            CALL g_properties.addAttribute(ls_name,ls_value)
            LET ls_name = adzp150_create_name(li_page, "detail_allow_insert_desc", "<<<")
            LET ls_value = "#此段落由產生器控制, 手動之設定無效!"
            CALL g_properties.addAttribute(ls_name,ls_value)
         ELSE
            LET ls_tmp = adzp150_create_name(li_page, "detail_var_title", "<<<")
            LET ls_tmp = g_properties.getValue(ls_tmp)
            LET ls_value = "IF g_insert = 'Y' AND NOT cl_null(g_insert) THEN \n",
                           "              CALL FGL_SET_ARR_CURR(",ls_tmp,".getLength()+1) \n",
                           "              LET g_insert = 'N' \n",
                           "           END IF \n"
            LET ls_name = adzp150_create_name(li_page, "detail_page_input", "<<<")
            CALL g_properties.addAttribute(ls_name,ls_value)
            #LET ls_value = "l_allow_insert"
            #LET ls_name = adzp150_create_name(li_page, "detail_allow_insert", "<<<")
            #CALL g_properties.addAttribute(ls_name,ls_value)
            LET ls_name = adzp150_create_name(li_page, "detail_allow_insert_desc", "<<<")
            LET ls_value = "#此頁面insert功能由產生器控制, 手動之設定無效!"
            CALL g_properties.addAttribute(ls_name,ls_value)
         END IF
         LET g_detail_page_info[li_page].id = g_detail_table_info[li_idx].id
         LET g_detail_page_info[li_page].fk = g_detail_table_info[li_idx].fk
         LET li_tmp = li_tmp + 1
      END WHILE     
   END FOR
   
   LET li_cnt = g_properties.getValue("page")
   FOR li_idx = 1 TO li_cnt
   
      LET ls_tmp = adzp150_create_name(li_idx, "detail_tbl_name", "<<<")
   
      #detail_fields_insert
      LET ls_tmp = adzp150_create_name(li_idx, "detail_fields_all", "<<<") 
      LET ls_value = g_properties.getValue(ls_tmp)
      LET lst_token = base.StringTokenizer.create(ls_value,",")
      LET ls_value = ""
      LET ls_tmp = g_detail_page_info[li_idx].id
      LET ls_tmp = ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      LET ls_tmp = g_detail_table_info[li_idx].pk

      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         IF adzp150_chk_intbl(g_detail_page_info[li_idx].id,ls_token) THEN
            IF ls_token.getIndexOf("(",1) THEN
               LET ls_token = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            END IF
            LET ls_value = ls_value, ls_token, ","
         END IF
      END WHILE
      LET ls_value = ls_value.subString(1,ls_value.getLength()-1)
      LET ls_name = adzp150_create_name(li_idx, "detail_fields_insert", "<<<") 
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #detail_fields_update
      IF NOT cl_null(g_detail_page_info[li_idx].fk)   AND
         NOT (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") AND
         (g_properties.getValue("type_id") <> "i07" OR g_properties.getValue("type_id") <> "i12") THEN
         LET ls_value = g_detail_page_info[li_idx].fk,",",ls_value
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_fields_update", "<<<") 
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #detail_vars_insert
      LET ls_tmp = adzp150_create_name(li_idx, "detail_vars_all", "<<<")
      LET ls_value = g_properties.getValue(ls_tmp)

      LET lst_token = base.StringTokenizer.create(ls_value,",")
      LET ls_value = ""
      LET ls_tmp = g_detail_page_info[li_idx].id
      LET ls_tmp = ".",ls_tmp.subString(1,ls_tmp.getIndexOf("_",1)-1)
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         IF adzp150_chk_intbl(g_detail_page_info[li_idx].id,ls_token) THEN
            LET ls_value = ls_value, ls_token, ","
         END IF
      END WHILE
      LET ls_value = ls_value.subString(1,ls_value.getLength()-1)
      LET ls_name = adzp150_create_name(li_idx, "detail_vars_insert", "<<<") 
      CALL g_properties.addAttribute(ls_name, ls_value)

      
      #detail_vars_update
      IF NOT cl_null(g_properties.getValue("master_var_allkeys")) AND 
         (g_properties.getValue("type_id") <> "i07" OR g_properties.getValue("type_id") <> "i12") THEN
         LET ls_value = g_detail_page_info[li_idx].var_fk,",",ls_value
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_vars_update", "<<<")
      CALL g_properties.addAttribute(ls_name, ls_value)


   END FOR
   
END FUNCTION

#+ 取得變數名稱
PUBLIC FUNCTION adzp150_create_name(pi_idx,ps_name,ps_type)
   DEFINE pi_idx     STRING
   DEFINE ps_name    STRING
   DEFINE ps_type    STRING
   DEFINE ls_return  STRING
   
   IF pi_idx = "1" AND ps_type = "<<<" THEN
      LET ls_return = ps_name
   ELSE
      LET ls_return = ps_name, pi_idx USING ps_type
   END IF
   
   RETURN ls_return

END FUNCTION

#+ 取得變數名稱
PRIVATE FUNCTION adzp150_detail_Filter(ps_input,pi_type)
   DEFINE ps_input     STRING
   DEFINE pi_type      SMALLINT   #精確的判斷欄位TRUE/FALSE
   DEFINE lst_token    base.StringTokenizer
   DEFINE ls_token     STRING
   DEFINE ls_tmp       STRING
   DEFINE ls_return    STRING

   LET lst_token = base.StringTokenizer.create(ps_input,",")
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      #濾除重複資料
      LET ls_tmp = ls_token.subString(ls_token.getIndexOf('.',1)+1,ls_token.getLength()), ","
      IF ls_return.getIndexOf(ls_tmp,1) = 0 THEN
         LET ls_return = ls_return, ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_return = ls_return, ","
         END IF
      ELSE
         IF NOT pi_type THEN  #不精確的狀況進來 #230112
            #有碰到原字串ls_return已有,l_xmdc055,然後 token(ls_tmp)要查,xmdc055,
            LET ls_tmp = ",",ls_tmp
            IF ls_return.getIndexOf(ls_tmp,1) = 0 THEN
               LET ls_return = ls_return, ls_token
               IF lst_token.hasMoreTokens() THEN
                  LET ls_return = ls_return, ","
               END IF
            END IF
         END IF
      END IF
   END WHILE
   
   IF ls_return.subString(ls_return.getLength(),ls_return.getLength()) = ',' THEN
      LET ls_return = ls_return.subString(1,ls_return.getLength()-1)
   END IF

   RETURN ls_return

END FUNCTION 
         
#+ 將table為主的資料轉為page為主的資料
PUBLIC FUNCTION adzp150_create_advanced_page_info()
   DEFINE li_cnt       INTEGER
   DEFINE li_idx       INTEGER
   DEFINE li_idx2      INTEGER
   DEFINE li_tmp       INTEGER
   DEFINE ls_tmp       STRING
   DEFINE ls_name      STRING
   DEFINE ls_value     STRING
   DEFINE lst_page     base.StringTokenizer
   DEFINE ls_page      STRING
   DEFINE lst_token    base.StringTokenizer
   DEFINE ls_token     STRING
   DEFINE lst_token2   base.StringTokenizer
   DEFINE ls_token2    STRING
   
   LET li_cnt = g_detail_table_info.getLength()
   FOR li_idx = 1 TO li_cnt
      LET lst_page = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      WHILE lst_page.hasMoreTokens()
         LET ls_page = lst_page.nextToken()
         
         #fisrt page
         IF g_detail_table_info[li_idx].first_page = ls_page THEN
            LET g_detail_page_info[ls_page].first_page = "Y"
         ELSE
            LET g_detail_page_info[ls_page].first_page = "N"
         END IF
         
         #detail_pk_num
         LET ls_tmp = adzp150_create_name(li_idx,"detail_pk_num_by_tbl", "<<<")
         LET ls_value = g_properties.getValue(ls_tmp)
         LET ls_name = adzp150_create_name(ls_page,"detail_pk_num", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_append_wc
         LET ls_tmp = adzp150_create_name(li_idx,"detail_append_wc_by_tbl", "<<<")
         LET ls_value = g_properties.getValue(ls_tmp)
         LET ls_name = adzp150_create_name(ls_page,"detail_append_wc", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_append_wc_s
         LET ls_tmp = adzp150_create_name(li_idx,"detail_append_wc_s_by_tbl", "<<<")
         LET ls_value = g_properties.getValue(ls_tmp)
         LET ls_name = adzp150_create_name(ls_page,"detail_append_wc_s", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_field_append
         LET ls_tmp = adzp150_create_name(li_idx,"detail_field_append_by_tbl", "<<<")
         LET ls_value = g_properties.getValue(ls_tmp)
         LET ls_name = adzp150_create_name(ls_page,"detail_field_append", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)

         #detail_var_append
         LET ls_tmp = adzp150_create_name(li_idx,"detail_var_append_by_tbl", "<<<")
         LET ls_value = g_properties.getValue(ls_tmp)
         LET ls_name = adzp150_create_name(ls_page,"detail_var_append", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_field_fks
         LET ls_name = adzp150_create_name(ls_page,"detail_field_fks", "<<<")
         LET ls_value = g_detail_table_info[li_idx].fk
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_var_fks
         #LET ls_value = g_properties.getValue("master_var_allkeys")
         #LET ls_name = adzp150_create_name(ls_page,"detail_var_fks", "<<<")
         #CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_field_pks
         LET ls_name = adzp150_create_name(ls_page,"detail_field_pks", "<<<")
         LET ls_value = g_detail_table_info[li_idx].pk
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_var_pks
         LET ls_tmp = adzp150_create_name(ls_page,"detail_field_pks", "<<<")
         LET ls_value = ""
         LET lst_token = base.StringTokenizer.create(g_properties.getValue(ls_tmp) , ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_tmp = adzp150_create_name(ls_page,"detail_var_title", "<<<")
            LET ls_tmp = g_properties.getValue(ls_tmp)
            LET ls_value = ls_value, ls_tmp, g_detail_table_info[li_idx].curr_idx, ls_token
            IF lst_token.hasMoreTokens() THEN
               LET ls_value = ls_value, "," 
            END IF
         END WHILE
         LET ls_name = adzp150_create_name(ls_page,"detail_var_pks", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #master_var_pks(for t02專用)
         IF ls_page = 1 AND (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
            LET ls_tmp = adzp150_create_name(ls_page,"detail_field_pks", "<<<")
            LET ls_value = ""
            LET lst_token = base.StringTokenizer.create(g_properties.getValue(ls_tmp) , ',')
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               LET ls_value = ls_value, "g_master.", ls_token
               IF lst_token.hasMoreTokens() THEN
                  LET ls_value = ls_value, "," 
               END IF
            END WHILE
            LET ls_name = adzp150_create_name(ls_page,"master_var_pks", "<<<")
            CALL g_properties.addAttribute(ls_name, ls_value)
         END IF
         
         #detail_page_bcl
         LET ls_tmp = adzp150_create_name(li_idx,"detail_table_bcl", "<<<")
         LET ls_value = g_properties.getValue(ls_tmp)
         LET ls_name = adzp150_create_name(ls_page,"detail_page_bcl", "<<<")
         CALL g_properties.addAttribute(ls_name, ls_value)

         #${detail${page}_field_fk${key}}
         LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].fk , ',')
         LET li_tmp = 1
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_name = adzp150_create_name(ls_page,"detail", "<<<")
            LET ls_name = adzp150_create_name(li_tmp,ls_name||"_field_fk","&&")
            LET ls_value = ls_token
            CALL g_properties.addAttribute(ls_name, ls_value)
            LET li_tmp = li_tmp + 1
         END WHILE
         
         #${detail${page}_field_pk${key}}
         LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].pk , ',')
         LET li_tmp = 1
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_name = adzp150_create_name(ls_page,"detail", "<<<")
            LET ls_name = adzp150_create_name(li_tmp,ls_name||"_field_pk","&&")
            LET ls_value = ls_token
            CALL g_properties.addAttribute(ls_name, ls_value)
            
            #特殊處理for t02
            IF ls_page = 1 AND (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
               LET ls_name = adzp150_create_name(li_tmp,"master_field_pk","&&")
               CALL g_properties.addAttribute(ls_name, ls_value)
            END IF
            
            LET li_tmp = li_tmp + 1
  
         END WHILE
         
         #${detail${page}_var_pk${key}}
         LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].pk , ',')
         LET li_tmp = 1
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_name = adzp150_create_name(ls_page,"detail", "<<<")
            LET ls_name = adzp150_create_name(li_tmp,ls_name||"_var_pk","&&")
            LET ls_tmp = adzp150_create_name(ls_page,"detail_var_title", "<<<")
            LET ls_value = g_properties.getValue(ls_tmp),g_detail_table_info[li_idx].curr_idx,ls_token
            CALL g_properties.addAttribute(ls_name, ls_value)
            #特殊處理for t02
            IF ls_page = 1 AND (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
               LET ls_name = adzp150_create_name(li_tmp,"master_var_pk","&&")
               LET ls_value = "g_master.",ls_token
               CALL g_properties.addAttribute(ls_name, ls_value)
            END IF
            LET li_tmp = li_tmp + 1
         END WHILE
         
         #${detail_tbl_name${page}} 
         LET ls_name = adzp150_create_name(ls_page,"detail_tbl_name", "<<<")
         LET ls_value = g_detail_table_info[li_idx].id
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_ins_Sync
         LET ls_value = 15 SPACES, "INITIALIZE gs_keys TO NULL \n"
         LET lst_token = base.StringTokenizer.create(g_detail_page_info[ls_page].var_fk , ',')
         LET li_tmp = 1
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_value = ls_value, 15 SPACES, "LET gs_keys[",li_tmp USING "<<<","] = ", 
                                     ls_token, "\n"
            LET li_tmp = li_tmp + 1
         END WHILE
         LET lst_token = base.StringTokenizer.create(g_detail_page_info[ls_page].var_pk , ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken() 
            LET ls_value = ls_value, 15 SPACES, "LET gs_keys[",li_tmp USING "<<<","] = ", ls_token, "\n"
            LET li_tmp = li_tmp + 1
         END WHILE
         LET ls_name = adzp150_create_name(ls_page, "detail_ins_Sync_pre", "<<<" ) 
         CALL g_properties.addAttribute(ls_name, ls_value)
         LET ls_value = ls_value, 15 SPACES, "CALL ", g_properties.getValue("general_prefix"), "_insert_b('",g_detail_page_info[ls_page].id,"',gs_keys,\"'",ls_page,"'\")"
         LET ls_name = adzp150_create_name(ls_page, "detail_ins_Sync", "<<<" ) 
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_upd_Sync
         LET ls_value = 15 SPACES, "INITIALIZE gs_keys TO NULL \n"
         LET lst_token = base.StringTokenizer.create(g_detail_page_info[ls_page].var_fk , ',')
         LET lst_token2 = base.StringTokenizer.create(g_detail_page_info[ls_page].var_fk_t , ',')
         LET li_tmp = 1
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_token2 = lst_token2.nextToken()
            LET ls_value = ls_value, 15 SPACES, 
                           "LET gs_keys[",    li_tmp USING "<<<","] = ", ls_token, "\n",
                           15 SPACES,
                           "LET gs_keys_bak[",li_tmp USING "<<<","] = ", ls_token2,"\n"
            LET li_tmp = li_tmp + 1
         END WHILE
         LET lst_token = base.StringTokenizer.create(g_detail_page_info[ls_page].var_pk , ',')
         LET lst_token2 = base.StringTokenizer.create(g_detail_page_info[ls_page].var_pk_t , ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_token2 = lst_token2.nextToken()
            LET ls_value = ls_value, 15 SPACES, 
                           "LET gs_keys[",    li_tmp USING "<<<","] = ", ls_token, "\n",
                           15 SPACES,
                           "LET gs_keys_bak[",li_tmp USING "<<<","] = ", ls_token2,"\n"
            LET li_tmp = li_tmp + 1
         END WHILE
         LET ls_value = ls_value, 15 SPACES, "CALL ", g_properties.getValue("general_prefix"), "_update_b('",g_detail_page_info[ls_page].id,"',gs_keys,gs_keys_bak,\"'",ls_page,"'\")"
         LET ls_name = adzp150_create_name(ls_page, "detail_upd_Sync", "<<<" ) 
         CALL g_properties.addAttribute(ls_name, ls_value)
                 
         #detail_del_Sync_pre
         LET ls_value = 15 SPACES, "INITIALIZE gs_keys TO NULL \n"
         LET lst_token = base.StringTokenizer.create(g_detail_page_info[ls_page].var_fk , ',')
         LET li_tmp = 1
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_value = ls_value, 15 SPACES, "LET gs_keys[",li_tmp USING "<<<","] = ", 
                                     ls_token, "\n"
            LET li_tmp = li_tmp + 1
         END WHILE
         LET lst_token = base.StringTokenizer.create(g_detail_page_info[ls_page].var_pk_t , ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken() 
            LET ls_value = ls_value, 15 SPACES, "LET gs_keys[",li_tmp USING "<<<","] = ", ls_token, "\n"
            LET li_tmp = li_tmp + 1
         END WHILE
         LET ls_name = adzp150_create_name(ls_page, "detail_del_Sync_pre", "<<<" ) 
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_del_Sync_post
         LET ls_value = 15 SPACES, "CALL ", g_properties.getValue("general_prefix"), "_delete_b('",g_detail_page_info[ls_page].id,"',gs_keys,\"'",ls_page,"'\")"
         LET ls_name = adzp150_create_name(ls_page, "detail_del_Sync_post", "<<<" ) 
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #detail_fields_default
         LET ls_value = g_detail_page_info[ls_page].default
         LET ls_name = adzp150_create_name(ls_page, "detail_fields_default", "<<<" ) 
         CALL g_properties.addAttribute(ls_name, ls_value)
         
         #create detail_idx_group
         LET li_tmp = 1
         LET ls_value = ""
         FOR li_tmp = 1 TO g_table_group.getLength()
            LET ls_value = g_table_group[li_tmp].page    	
            IF ls_value.getIndexOf(ls_page||",",1) > 0 THEN
               LET ls_name = adzp150_create_name(li_idx, "detail_idx_group", "<<<")
               LET ls_value = g_table_group[li_tmp].page2
               CALL g_properties.addAttribute(ls_name, ls_value)
               EXIT FOR
            END IF
         END FOR 
		 
      END WHILE
   
   END FOR
   
END FUNCTION

#+ 多語言處理
PRIVATE FUNCTION adzp150_create_lang()
   DEFINE li_cnt  INTEGER
   DEFINE li_idx  INTEGER
   DEFINE li_idx2 INTEGER
   DEFINE ls_pk   STRING
   
   LET li_cnt = g_detail_lang_info.getLength()
   #LET li_idx2 = g_detail_table_info.getLength()
   #WHILE TRUE
   #   LET ls_pk = g_detail_table_info[li_idx2].pk
   #   IF NOT cl_null(ls_pk) THEN
   #      EXIT WHILE
   #   END IF
   #   LET li_idx2 = li_idx2 - 1
   #END WHILE
   FOR li_idx = 1 TO li_cnt
      #160617 
      CALL adzp150_detail_multi_table(g_detail_lang_info[li_idx].node,
                                      g_detail_lang_info[li_idx].master,
                                      g_detail_lang_info[li_idx].master_pk,
                                      g_detail_lang_info[li_idx].master_idx,
                                      g_detail_lang_info[li_idx].master_fk,
                                      g_detail_lang_info[li_idx].master_single_fk)
      #CALL adzp150_update_item_action('d',g_detail_lang_info[li_idx].node)   
   END FOR

END FUNCTION 


#+ 增加key欄位檢查段落
PRIVATE FUNCTION adzp150_create_chk()
   DEFINE li_cnt       INTEGER
   DEFINE li_idx       INTEGER
   DEFINE ls_page      STRING
   DEFINE ls_name      STRING
   DEFINE ls_tmp       STRING
   DEFINE ls_point     STRING
   DEFINE lst_token    base.StringTokenizer
   DEFINE ls_token     STRING
   
   LET li_cnt = g_detail_page_info.getLength()
   FOR li_idx = 1 TO li_cnt 
   
      LET ls_point = adzp150_multi_table_key_field_chk(li_idx)
      
      LET ls_tmp = adzp150_create_name(li_idx, "detail_field_pks", "<<<")
      
      LET ls_tmp = g_properties.getValue(ls_tmp)
      
      LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
      
         IF g_detail_page_info[li_idx].first_page = "N" THEN
            LET ls_name = ls_token, "_", li_idx USING "<<<"
         ELSE
            LET ls_name = ls_token
         END IF
         
         LET ls_page = "page", adzp150_page_trans(li_idx), "." 
         
         #寫進DB
         LET g_dzbb.prog_name   = g_properties.getValue("app_id")
         LET g_dzbb.point_name  = "input.a.",ls_page,ls_name
         LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
         LET g_dzbb.addpoint    = ls_point
         CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
      END WHILE
   END FOR

END FUNCTION

#+ 增加key欄位檢查段落
PRIVATE FUNCTION adzp150_multi_table_key_field_chk(ps_page)
   DEFINE ps_page              STRING
   DEFINE li_lv                INTEGER
   DEFINE li_space             INTEGER 
   DEFINE ls_var               STRING               #變數名稱
   DEFINE ls_sql               STRING               #暫存sql
   DEFINE ls_filled_chk        STRING               #檢查所有欄位都已填
   DEFINE ls_insert_chk        STRING               #確定是否為新增段
   DEFINE ls_modify_chk        STRING               #確定是否為修改段
   DEFINE ls_return            STRING               
   DEFINE ls_tmp               STRING               
   DEFINE ls_name              STRING               #欄位名稱
   DEFINE ls_key_fileds        STRING               #單身所有key欄位
   DEFINE ls_key_vars          STRING               #單身所有key變數
   DEFINE ls_key_vars_tmp      STRING               #單身所有key變數(bak)
   DEFINE lb_tok               base.StringTokenizer
   DEFINE lb_tok2              base.StringTokenizer
   DEFINE ls_field             STRING
   DEFINE li_idx               INTEGER
   
   LET li_lv  = 3
   LET li_space = 3
   
   #key變數資料
   LET ls_key_vars = g_detail_page_info[ps_page].var_fk,',',
                     g_detail_page_info[ps_page].var_pk

   IF (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") AND g_properties.getValue(adzp150_create_name(ps_page, "general_detail_mark", "<<<")) = "#" THEN
      LET ls_tmp = g_properties.getValue("detail_var_pks")
      LET ls_tmp = cl_replace_str(ls_tmp, '[l_ac]', '[g_detail_idx]')
      LET ls_key_vars = ls_tmp,",",
                        g_properties.getValue(adzp150_create_name(ps_page, "detail_var_pks", "<<<"))
   END IF
   
   #key欄位資料
   LET ls_key_fileds = g_properties.getValue(adzp150_create_name(ps_page, "detail_field_fks", "<<<")),",",
                       g_properties.getValue(adzp150_create_name(ps_page, "detail_field_pks", "<<<"))
    
   #key變數資料(bak)
   LET lb_tok = base.StringTokenizer.create(g_properties.getValue("master_field_allkeys"),",")
   IF (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") AND g_properties.getValue(adzp150_create_name(ps_page, "general_detail_mark", "<<<")) = "#" THEN
      LET lb_tok = base.StringTokenizer.create(g_properties.getValue("detail_field_allkeys"),",")
   END IF
   
   LET lb_tok = base.StringTokenizer.create(g_detail_page_info[ps_page].var_fk_t,",")
   WHILE lb_tok.hasMoreTokens()
      LET ls_tmp = lb_tok.nextToken()
      LET ls_name = adzp150_create_name(ps_page, "detail_var_title", "<<<")
      IF (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") AND g_properties.getValue(adzp150_create_name(ps_page, "general_detail_mark", "<<<")) = "#" THEN
         LET ls_name = adzp150_create_name(1, "detail_var_title", "<<<")
         LET ls_key_vars_tmp = ls_key_vars_tmp,ls_tmp,","
      ELSE
         LET ls_key_vars_tmp = ls_key_vars_tmp,ls_tmp,","
      END IF
   END WHILE

   LET lb_tok = base.StringTokenizer.create(g_detail_page_info[ps_page].var_pk_t,",")
   WHILE lb_tok.hasMoreTokens()
      LET ls_tmp = lb_tok.nextToken()
      LET ls_key_vars_tmp = ls_key_vars_tmp,ls_tmp,","
   END WHILE
   
   #檢查所有單身欄位都已填(ls_filled_chk)
   LET ls_filled_chk = ""
   
   LET lb_tok = base.StringTokenizer.create(ls_key_vars,",")
   WHILE lb_tok.hasMoreTokens()
      LET ls_tmp = lb_tok.nextToken()
      LET ls_filled_chk = ls_filled_chk," ",ls_tmp," IS NOT NULL "
      IF lb_tok.hasMoreTokens() THEN
         LET ls_filled_chk = ls_filled_chk,"AND"
      END IF
   END WHILE

   #確定是否為新增段(ls_insert_chk)
   LET ls_insert_chk = " l_cmd = 'a' "  

   #確定是否為修改段(ls_modify_chk)
   LET ls_modify_chk = " l_cmd = 'u' AND ("

   LET lb_tok = base.StringTokenizer.create(ls_key_vars_tmp,",")
   LET lb_tok2 = base.StringTokenizer.create(ls_key_vars,",")
   WHILE lb_tok.hasMoreTokens()
      LET ls_field = lb_tok.nextToken()
      LET ls_var = lb_tok2.nextToken()
      LET ls_modify_chk = ls_modify_chk, ls_var, " != ",ls_field
      IF lb_tok.hasMoreTokens() THEN
         LET ls_modify_chk = ls_modify_chk, " OR "
      END IF
   END WHILE
   LET ls_modify_chk = ls_modify_chk,")"

   #組合notDup的SQL(ls_sql)
   LET ls_name = adzp150_create_name(ps_page, "detail_tbl_name", "<<<")
   LET ls_sql = "\"SELECT COUNT(1) FROM ",g_properties.getValue(ls_name)," WHERE \"||"
                                        
   #若有ent,site code則須加額外判斷
   LET ls_name = adzp150_create_name(ps_page, "detail_append_wc_s", "<<<")
   IF NOT cl_null(g_properties.getValue(ls_name)) THEN
      LET ls_sql = ls_sql,"\"",g_properties.getValue(ls_name) ,"\"||" 
   END IF

   #判斷所有單身key for ls_sql
   LET lb_tok = base.StringTokenizer.create(ls_key_fileds,",")
   LET lb_tok2 = base.StringTokenizer.create(ls_key_vars,",")
   WHILE lb_tok.hasMoreTokens()
      LET ls_field = lb_tok.nextToken()
      LET ls_var = lb_tok2.nextToken()

      LET ls_sql = ls_sql,"\"",ls_field," = '\"||",ls_var
      IF lb_tok.hasMoreTokens() THEN
         LET ls_sql = ls_sql, " ||\"' AND \"|| "
      ELSE
         LET ls_sql = ls_sql, " ||\"'\""
      END IF
   END WHILE

   LET g_value = ls_filled_chk
   CALL g_properties.addAttribute("mdl_filled_chk",g_value)
   LET g_value = ls_modify_chk
   CALL g_properties.addAttribute("mdl_modify_chk",g_value)
   LET g_value = ls_sql
   CALL g_properties.addAttribute("mdl_sql",g_value)
   LET g_value = "l_cmd"
   CALL g_properties.addAttribute("mdl_loc",g_value)
   LET ls_return = ls_return, adzp150_make_slice("a05")

   RETURN ls_return
   
END FUNCTION


#+ 濾除非必要欄位_X
PUBLIC FUNCTION adzp150_field_filter(ps_string,ps_type)
   DEFINE ps_string        STRING
   DEFINE ps_type          STRING
   DEFINE ls_return        STRING
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token         STRING
   DEFINE ls_field         STRING
   DEFINE ls_field_list    STRING
   DEFINE ls_field_type    STRING

   CASE ps_type
      WHEN "field" 
         LET lst_token = base.StringTokenizer.create(ps_string, ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_field_type = ""
            #曾因不明原因註解
            IF ls_token.getIndexOf("_",1)         AND 
               NOT ls_token.getIndexOf("_desc",1) AND
               NOT ls_token.getIndexOf("l_",1)    AND
               NOT ls_token.getIndexOf("lc_",1)   AND
               NOT ls_token.getIndexOf("(",ls_token.getIndexOf("_",1)+1) = 0 THEN
               LET ls_field = ls_token.subString(1, ls_token.getIndexOf("_",1)-1)
               #IF ls_token.getIndexOf("(",1) > 0 THEN
               #   LET ls_field_type = ls_token.subString(ls_token.getIndexOf("(",1),ls_token.getIndexOf(")",1))
               #END IF
               LET ls_return = ls_return, ls_field, ls_field_type, ","
            ELSE
               LET ls_return = ls_return, ls_token, ","
            END IF
         END WHILE
         LET ls_return = ls_return.subString(1, ls_return.getLength()-1)
         
      WHEN "sql" 
         LET ls_field_list = ps_string.subString(ps_string.getIndexOf("SELECT",1)+7,
                                                 ps_string.getIndexOf("FROM",1)-2  )
         LET lst_token = base.StringTokenizer.create(ls_field_list, ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            IF ls_token.getIndexOf("_",1) THEN
               LET ls_field = ls_token.subString(1, ls_token.getIndexOf("_",1)-1)
               LET ls_return = ls_return, ls_field, ","
            ELSE
               LET ls_return = ls_return, ls_token, ","
            END IF
         END WHILE
         LET ls_return = ls_return.subString(1, ls_return.getLength()-1)
         
         LET ls_return = cl_replace_str(ps_string,ls_field_list,ls_return)
   END CASE

   RETURN ls_return
   
END FUNCTION

#+ 單身key欄位發生異動(for 多單身)
PUBLIC FUNCTION adzp150_detail_multi_table_detail_key_upd()
   DEFINE lst_token      base.StringTokenizer
   DEFINE ls_token       STRING
   DEFINE lst_token2     base.StringTokenizer
   DEFINE ls_token2      STRING
   DEFINE lst_token3     base.StringTokenizer
   DEFINE ls_token3      STRING
   DEFINE li_cnt         INTEGER 
   DEFINE li_idx         INTEGER 
   DEFINE li_tmp         INTEGER 
   DEFINE li_tmp2        INTEGER 
   DEFINE ls_name        STRING
   DEFINE ls_value       STRING
     
   LET li_cnt = g_detail_table_info.getLength()
   FOR li_idx = 1 TO li_cnt
      #detail_append_pk_upd_by_tbl 單筆修改(by table)   
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_append_pk_upd_tmp", "<<<")  
         LET ls_value = ls_value, g_properties.getValue(ls_name)
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_append_pk_upd_by_tbl", "<<<")  
      CALL g_properties.addAttribute(ls_name, ls_value)

      #detail_append_pk_upd_all_by_tbl 多筆修改(by table)
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()         
         LET ls_name = adzp150_create_name(ls_token, "detail_append_pk_upd_all_tmp", "<<<")  
         LET ls_value = ls_value, g_properties.getValue(ls_name)
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_append_pk_upd_all_by_tbl", "<<<")  
      CALL g_properties.addAttribute(ls_name, ls_value)
      
      #160617
      #detail_append_pk_del_all_by_tbl 多筆刪除(by table)
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()         
         LET ls_name = adzp150_create_name(ls_token, "detail_append_pk_del_all_tmp", "<<<")  
         LET ls_value = ls_value, g_properties.getValue(ls_name)
      END WHILE
      LET ls_name = adzp150_create_name(li_idx, "detail_append_pk_del_all_by_tbl", "<<<")  
      CALL g_properties.addAttribute(ls_name, ls_value)
       
   END FOR

END FUNCTION

#+ 單身link特殊處理
PRIVATE FUNCTION adzp150_create_page_link_info()
   DEFINE li_page           INTEGER
   DEFINE ls_page           STRING
   DEFINE li_table_group    INTEGER
   DEFINE li_table          INTEGER
   DEFINE li_idx            INTEGER
   DEFINE li_idx2           INTEGER
   DEFINE ls_pages          STRING
   DEFINE ls_first_page     STRING
   DEFINE ls_name           STRING
   DEFINE ls_value          STRING
   
   LET li_page = g_properties.getValue("page")
   LET li_table_group = g_table_group.getLength()
   LET li_table = g_detail_table_info.getLength()
   
   FOR li_idx = 1 TO li_page
      LET ls_page = li_idx USING "<<<"
      FOR li_idx2 = 1 TO li_table_group
         LET ls_pages = g_table_group[li_idx2].page
         IF ls_pages.getIndexOf(ls_page,1) > 0 THEN
            IF ls_pages.getIndexOf(ls_page,1) = 1 THEN
               LET g_detail_page_link_info[li_idx].page = 0
            ELSE
               LET ls_first_page = ls_pages.subString(1,ls_pages.getIndexOf(',',1)-1)
               LET g_detail_page_link_info[li_idx].page = ls_first_page
            END IF
         END IF
      END FOR
      
      FOR li_idx2 = 1 TO li_table
         LET ls_pages = g_detail_table_info[li_idx2].page
         #串聯的table
         IF ls_pages.getIndexOf(g_detail_page_link_info[li_idx].page,1) > 0 THEN
            LET g_detail_page_link_info[li_idx].table = g_detail_table_info[li_idx2].id
            LET ls_name = adzp150_create_name(li_idx2, "detail_by_tbl_var_pks", "<<<")
            LET g_detail_page_link_info[li_idx].var_keys = g_properties.getValue(ls_name)
         END IF
      END FOR
      
      LET ls_name = adzp150_create_name(li_idx, "detail_page_link", "<<<")
      LET ls_value = g_detail_page_link_info[li_idx].page
      CALL g_properties.addAttribute(ls_name,ls_value)
          
      LET ls_name = adzp150_create_name(li_idx, "detail_page_link", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(ls_value, "detail_var_title", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_idx, "detail_link_title", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
          
      LET ls_name = adzp150_create_name(li_idx, "detail_link_key", "<<<")
      LET ls_value = g_detail_page_link_info[li_idx].var_keys
      LET ls_value = cl_replace_str(ls_value,'g_detail_idx2','l_ac')
      LET ls_value = cl_replace_str(ls_value,'g_detail_idx' ,'l_ac')
      CALL g_properties.addAttribute(ls_name,ls_value)
   END FOR

END FUNCTION

#+ detail_vars_all重製(前製)
PRIVATE FUNCTION adzp150_create_detail_vars_all_pre()
   DEFINE ls_name          STRING
   DEFINE ls_value         STRING
   DEFINE li_idx           INTEGER
   DEFINE ls_pages         STRING
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token         STRING
   DEFINE ls_allvar        STRING
   DEFINE la_allvar        DYNAMIC ARRAY OF STRING #存放每個page的vars
   DEFINE ls_tbl_name      STRING 
   DEFINE ls_var_title     STRING 

   #先組出對應的vars_all(使用fields_all處理)
   FOR li_idx = 1 TO g_properties.getValue("page")
      LET ls_name = adzp150_create_name(li_idx, "detail_tbl_name", "<<<") 
      LET ls_tbl_name = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_idx, "detail_var_title", "<<<") 
      LET ls_var_title = g_properties.getValue(ls_name),'[l_ac].'
      LET ls_name = adzp150_create_name(li_idx, "detail_fields_all", "<<<") 
      LET lst_token = base.StringTokenizer.create(g_properties.getValue(ls_name), ',')

      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         IF ls_token.getIndexOf('(',1) THEN
            LET ls_token = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
         END IF
         IF cl_getField(ls_tbl_name,ls_token) THEN
            LET la_allvar[li_idx] = la_allvar[li_idx], ls_var_title,ls_token, ','
         END IF
      END WHILE
      LET la_allvar[li_idx] = la_allvar[li_idx].subString(1,la_allvar[li_idx].getLength()-1)
      LET ls_name = adzp150_create_name(li_idx, "detail_vars_all", "<<<")
      LET ls_value = la_allvar[li_idx]
      CALL g_properties.addAttribute(ls_name, ls_value)

   END FOR 

END FUNCTION

#+ detail_vars_all重製(後製)
PRIVATE FUNCTION adzp150_create_detail_vars_all()
   DEFINE ls_name          STRING
   DEFINE ls_value         STRING
   DEFINE li_idx           INTEGER
   DEFINE ls_pages         STRING
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token         STRING
   DEFINE ls_allvar        STRING
   DEFINE la_allvar        DYNAMIC ARRAY OF STRING #存放每個page的vars
   DEFINE ls_tbl_name      STRING 
   DEFINE ls_var_title     STRING 
   
   FOR li_idx = 1 TO g_detail_table_info.getLength()
      #重組detail_vars_all
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      LET ls_value = ''
      LET ls_allvar = ''
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_vars_all", "<<<")
         LET ls_value = g_properties.getValue(ls_name)
         IF NOT cl_null(ls_value) THEN
            LET ls_allvar = ls_allvar, ls_value, ','
         END IF
      END WHILE
      LET ls_allvar = ls_allvar.subString(1,ls_allvar.getLength()-1)
      
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_vars_all", "<<<") 
         LET ls_value = ls_allvar
         CALL g_properties.addAttribute(ls_name,ls_value)
      END WHILE
      
      #重組detail_vars_update
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      LET ls_token = lst_token.nextToken()
      LET ls_name = adzp150_create_name(ls_token, "detail_vars_update", "<<<") 
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_allvar = ls_value
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_vars_update", "<<<") 
         LET ls_value = g_properties.getValue(ls_name)
         LET ls_allvar = ls_allvar, ',', ls_value
      END WHILE
      
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_vars_update", "<<<") 
         LET ls_value = adzp150_detail_Filter(ls_allvar,TRUE)
         CALL g_properties.addAttribute(ls_name,ls_value)
      END WHILE
      
      #重組detail_fields_update
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      LET ls_token = lst_token.nextToken()
      LET ls_name = adzp150_create_name(ls_token, "detail_fields_update", "<<<") 
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_allvar = ls_value
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_fields_update", "<<<") 
         LET ls_value = g_properties.getValue(ls_name)
         LET ls_allvar = ls_allvar, ',', ls_value
      END WHILE
      
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_fields_update", "<<<") 
         LET ls_value = adzp150_detail_Filter(ls_allvar,TRUE)
         CALL g_properties.addAttribute(ls_name,ls_value)
      END WHILE
      
      #重組 detail_multi_table_delete (160627)
      LET ls_value = ""
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, "\n",
                        g_properties.getValue("detail_multi_table_delete"||ls_token)
      END WHILE
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = "detail_multi_table_delete",ls_token
         CALL g_properties.addAttribute(ls_name,ls_value)
         #DISPLAY ls_name,":",ls_value
      END WHILE
      
      #161130-00056 ---start---
      #重組 detail_multi_table_delete_before_delete
      LET ls_value = ""
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, "\n",
                        g_properties.getValue("detail_multi_table_delete_before_delete"||ls_token)
      END WHILE
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = "detail_multi_table_delete_before_delete",ls_token
         CALL g_properties.addAttribute(ls_name,ls_value)
         #DISPLAY ls_name,":",ls_value
      END WHILE

      #重組 detail_multi_table_delete_delete_function
      LET ls_value = ""
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, "\n",
                        g_properties.getValue("detail_multi_table_delete_delete_function"||ls_token)
      END WHILE
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = "detail_multi_table_delete_delete_function",ls_token
         CALL g_properties.addAttribute(ls_name,ls_value)
         #DISPLAY ls_name,":",ls_value
      END WHILE
      #161130-00056 --- end ---

      #重組 detail_multi_table_bak (160627)
      LET ls_value = ""
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, "\n",
                        g_properties.getValue(adzp150_create_name(ls_token, "detail_multi_table_bak", "<<<") )
      END WHILE
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = adzp150_create_name(ls_token, "detail_multi_table_bak", "<<<")
         CALL g_properties.addAttribute(ls_name,ls_value)
         #DISPLAY ls_name,":",ls_value
      END WHILE
      
      #重組 detail_multi_table_lock (160627)
      LET ls_value = ""
      LET ls_pages = g_detail_table_info[li_idx].page
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, "\n",
                        g_properties.getValue("detail_multi_table_lock"||ls_token)
      END WHILE
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = "detail_multi_table_lock",ls_token
         CALL g_properties.addAttribute(ls_name,ls_value)
         #DISPLAY ls_name,":",ls_value
      END WHILE
      
      
   END FOR
   
END FUNCTION

#+ 單身, 相關資料擷取 
FUNCTION adzp150_detail_info(p_node,pi_idx)
   DEFINE p_node            om.DomNode
   DEFINE pi_idx            INTEGER
   DEFINE li_idx            INTEGER
   DEFINE li_page_idx       INTEGER
   DEFINE ls_master         STRING
   DEFINE ls_master_pages   STRING
   DEFINE li_master         INTEGER
   DEFINE ls_name           STRING
   DEFINE ls_value          STRING
   DEFINE ls_tmp            STRING
   DEFINE lst_token         base.StringTokenizer
   DEFINE ls_token          STRING
   DEFINE lst_page_token    base.StringTokenizer
   DEFINE li_page_token     INTEGER
   DEFINE ls_delete         STRING
   DEFINE ls_page_reproduce STRING
   
   #mapping_fk
   LET g_detail_table_info[pi_idx].mapping_fk = g_properties.getValue("master_field_allkeys")
   
   #var_fk
   LET g_detail_table_info[pi_idx].var_fk = g_properties.getValue("master_var_allkeys")

   #var_fk_t
   LET lst_token = base.StringTokenizer.create(g_detail_table_info[pi_idx].var_fk , ',')
   LET li_idx = 1
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_tmp = adzp150_create_name(li_idx, "master_field_pk", "&&")
      LET ls_tmp = g_properties.getValue(ls_tmp)
      LET ls_value = ls_value, "g_", ls_tmp, "_t"
      IF lst_token.hasMoreTokens() THEN
         LET ls_value = ls_value, ','
      END IF
      LET li_idx = li_idx + 1
   END WHILE
    
   LET g_detail_table_info[pi_idx].var_fk_t = ls_value
            
   #var_pk_t
   LET lst_token = base.StringTokenizer.create(g_detail_table_info[pi_idx].pk , ',')
   LET ls_value = ""
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_tmp = adzp150_create_name(g_detail_table_info[pi_idx].first_page, "detail_var_title", "<<<")
      LET ls_tmp = g_properties.getValue(ls_tmp)
      LET ls_value = ls_value, ls_tmp, "_t.",ls_token
      IF lst_token.hasMoreTokens() THEN
         LET ls_value = ls_value, ','
      END IF
   END WHILE
    
   LET g_detail_table_info[pi_idx].var_pk_t = ls_value
   
   #var_pk
   LET ls_tmp = p_node.getAttribute("pk")    
   LET ls_value = ""
   LET lst_token = base.StringTokenizer.create(ls_tmp , ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_tmp = adzp150_create_name(g_detail_table_info[pi_idx].first_page,"detail_var_title", "<<<")
      LET ls_tmp = g_properties.getValue(ls_tmp)
      LET ls_value = ls_value, ls_tmp, "[g_detail_idx].", ls_token
      IF lst_token.hasMoreTokens() THEN
         LET ls_value = ls_value, "," 
      END IF
   END WHILE
   LET g_detail_table_info[pi_idx].var_pk = ls_value
   
   IF g_properties.getValue("type_id") = "t01" OR (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
      LET g_detail_table_info[pi_idx].curr_idx = "[g_detail_idx]."
   ELSE
      LET g_detail_table_info[pi_idx].curr_idx = "[l_ac]."
   END IF
   
   #自己的detail_fill_chk
   LET ls_name = adzp150_create_name(pi_idx, "detail_fill_chk", "<<<")
   LET ls_value = g_properties.getValue(ls_name)
   IF NOT cl_null(ls_value) THEN
      LET ls_value = ls_value, 'OR \n'
   END IF
   LET ls_value = "(NOT cl_null(g_wc2_table",pi_idx USING "<<<",") AND ",
                  "g_wc2_table",pi_idx USING "<<<",".trim() <> '1=1') "
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #detail_fill_chk_all
   LET ls_name = "detail_fill_chk_all"
   LET ls_value = g_properties.getValue(ls_name)
   IF NOT cl_null(ls_value) THEN
      LET ls_value = ls_value, ' AND \n      '
   END IF
   LET ls_value = ls_value, 
                  "(cl_null(g_wc2_table",pi_idx USING "<<<",") OR ",
                  "g_wc2_table",pi_idx USING "<<<",".trim() = '1=1') "
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   LET g_detail_table_info[pi_idx].type = "M"
   
   LET li_page_idx = 0
   LET lst_page_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
   WHILE lst_page_token.hasMoreTokens()
      LET li_page_token = lst_page_token.nextToken()
      LET li_page_idx = li_page_idx + 1

      #var_fk
      LET g_detail_page_info[li_page_token].var_fk = g_properties.getValue("master_var_allkeys")
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_fks", "<<<")
      LET ls_value = g_detail_page_info[li_page_token].var_fk      
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #var_fk_t
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[pi_idx].var_fk , ',')
      LET li_idx = 1
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(li_idx, "master_field_pk", "&&")
         LET ls_tmp = g_properties.getValue(ls_tmp)
         LET ls_value = ls_value, "g_", ls_tmp, "_t"
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ','
         END IF
         LET li_idx = li_idx + 1
      END WHILE
       
      LET g_detail_page_info[li_page_token].var_fk_t = ls_value

      #var_pk_t
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[pi_idx].pk , ',')
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(li_page_token, "detail_var_title", "<<<")
         LET ls_tmp = g_properties.getValue(ls_tmp)
         LET ls_value = ls_value, ls_tmp, "_t.",ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ','
         END IF
      END WHILE
       
      LET g_detail_page_info[li_page_token].var_pk_t = ls_value
      
      #var_pk
      LET ls_tmp = p_node.getAttribute("pk")    
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(ls_tmp , ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(li_page_token,"detail_var_title", "<<<")
         LET ls_tmp = g_properties.getValue(ls_tmp)
         LET ls_value = ls_value, ls_tmp, "[g_detail_idx].", ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, "," 
         END IF
      END WHILE
      LET g_detail_page_info[li_page_token].var_pk = ls_value
      
      #detail_page_delete_by_tbl(組合)
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_title", "<<<") 
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_delete = ls_delete, "      CALL ", ls_value,".deleteElement(",ls_value,".getLength())\n"
      
      #detail_first_field
      LET ls_name = "detail_page_field", li_page_token USING "<<<"
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_page_token, "detail_first_field", "<<<") 
      LET lst_token = base.StringTokenizer.create(ls_value, ',')
      LET ls_value = lst_token.nextToken()
     IF ls_value.getIndexOf('(',1) THEN
        LET ls_value = ls_value.subString(1,ls_value.getIndexOf('(',1)-1)
     END IF
      IF li_page_idx > 1 THEN
         LET ls_tmp = g_detail_table_info[pi_idx].pk
         #key要加上額外編號
         IF ls_tmp.getIndexOf(ls_value,1) > 0 THEN
            LET ls_value = ls_value, "_", li_page_token USING "<<<"
         END IF
      END IF
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_page_reproduce
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_title", "<<<") 
      LET ls_value = g_properties.getValue(ls_name)
      IF NOT cl_null(ls_page_reproduce) THEN
         LET ls_page_reproduce = ls_page_reproduce, "               "
      END IF
      LET ls_page_reproduce = ls_page_reproduce, 
                              "LET ",ls_value,"[li_reproduce_target].* = ",
                              ls_value,"[li_reproduce].*\n"
      
   END WHILE  
   
   #detail_page_reproduce(回填到每一個page)
   LET lst_page_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
   WHILE lst_page_token.hasMoreTokens()
      LET li_page_token = lst_page_token.nextToken() 
      LET ls_name = adzp150_create_name(li_page_token, "detail_page_reproduce", "<<<") 
      LET ls_value = ls_page_reproduce
      CALL g_properties.addAttribute(ls_name,ls_value)
   END WHILE
   
   #detail_page_delete_by_tbl
   LET ls_name = adzp150_create_name(pi_idx, "detail_page_delete_by_tbl", "<<<")  
   LET ls_value = ls_delete
   CALL g_properties.addAttribute(ls_name,ls_value)
   
END FUNCTION

#+ 單身的單身, 相關資料擷取 
FUNCTION adzp150_detail2_info(p_node,pi_idx)
   DEFINE p_node            om.DomNode
   DEFINE pi_idx            INTEGER
   DEFINE li_idx            INTEGER
   DEFINE li_page_idx       INTEGER
   DEFINE ls_master         STRING
   DEFINE ls_master_pages   STRING
   DEFINE li_master         INTEGER
   DEFINE li_master_group   INTEGER
   DEFINE ls_name           STRING
   DEFINE ls_value          STRING
   DEFINE ls_tmp            STRING
   DEFINE lc_tmp            CHAR(100)
   DEFINE lc_tmp2           CHAR(100)
   DEFINE li_tmp            INTEGER
   DEFINE lst_page_token    base.StringTokenizer
   DEFINE li_page_token     INTEGER
   DEFINE lst_token         base.StringTokenizer
   DEFINE ls_token          STRING
   DEFINE lst_token2        base.StringTokenizer
   DEFINE ls_token2         STRING
   DEFINE ls_clear          STRING
   DEFINE ls_delete         STRING
   DEFINE ls_page_reproduce STRING
   
   #取得所有第二層單身的第一個頁籤(detail_first_page,detail_first_page_name)
   IF g_properties.getValue("detail_first_page") IS NOT NULL THEN
      LET li_idx = g_properties.getValue("detail_first_page")
   ELSE
      LET li_idx = 999
   END IF
   
   LET lst_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
   WHILE lst_token.hasMoreTokens()
      LET li_page_idx = lst_token.nextToken()
      IF li_idx > li_page_idx THEN
         LET li_idx = li_page_idx
         LET ls_name  = "detail_first_page"
         LET ls_value = li_page_idx USING "<<<"
         CALL g_properties.addAttribute(ls_name,ls_value)
         LET ls_name  = "detail_first_page_name"
         LET ls_value = "detail_var_title", li_page_idx USING "<<<"
         LET ls_value = g_properties.getValue(ls_value)
         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF
   END WHILE
   LET li_idx = 0
   LET li_page_idx = 0
   
   #擷取上層table名稱
   LET ls_master = p_node.getAttribute("master")
   
   #首先找出這個單身的上層單身
   FOR li_idx = 1 TO g_detail_table_info.getLength()
      IF g_detail_table_info[li_idx].id = ls_master THEN
         LET li_master = li_idx
         EXIT FOR
      END IF
   END FOR
   
   #定義上層單身(編號)
   LET ls_name = adzp150_create_name(pi_idx, "detail_master_idx", "<<<")
   CALL g_properties.addAttribute(ls_name,li_master)
   
   #定義上層單身(名稱)
   LET ls_name = adzp150_create_name(pi_idx, "detail_master_name", "<<<")
   CALL g_properties.addAttribute(ls_name,ls_master)
   
   #若找不到則斷定為T02樣板且第一個table為上階
   IF li_master = 0 AND (g_properties.getValue("type_id") = "t02" OR g_properties.getValue("type_id") = "q02") THEN
      LET li_master = 1
   END IF
   
   #detail_append_join_detail
   LET ls_value = ""
   LET ls_name = adzp150_create_name(li_master, "detail_append_join_detail", "<<<")
   LET ls_value = g_properties.getValue(ls_name)
   IF NOT cl_null(ls_value) THEN
      LET ls_value = ls_value, '\n               '
   END IF
   LET ls_value = ls_value, "\" LEFT JOIN ", p_node.getAttribute("id"), " ON "
   
   LET lc_tmp2 = g_detail_table_info[li_master].id 
   #SELECT dzed004 INTO lc_tmp FROM dzed_t
   # WHERE dzed001 = lc_tmp2
   #   AND dzed003 = 'P'
   LET lc_tmp = g_detail_table_info[li_master].fk,',',g_detail_table_info[li_master].pk
   IF NOT cl_null(adzp150_chk_ent_and_site(g_detail_table_info[li_master].id,'site')) THEN
      LET lc_tmp = adzp150_chk_ent_and_site(g_detail_table_info[li_master].id,'site'),',',lc_tmp
   END IF
   IF NOT cl_null(adzp150_chk_ent_and_site(g_detail_table_info[li_master].id,'ent')) THEN
      LET lc_tmp = adzp150_chk_ent_and_site(g_detail_table_info[li_master].id,'ent'),',',lc_tmp
   END IF
   LET lst_token = base.StringTokenizer.create(lc_tmp, ',')

   LET lc_tmp2 = p_node.getAttribute("id")
   #SELECT dzed004 INTO lc_tmp FROM dzed_t
   # WHERE dzed001 = lc_tmp2
   #   AND dzed003 = 'P'
   LET lc_tmp = p_node.getAttribute("fk"),',',p_node.getAttribute("pk")
   IF NOT cl_null(adzp150_chk_ent_and_site(p_node.getAttribute("id"),'site')) THEN
      LET lc_tmp = adzp150_chk_ent_and_site(p_node.getAttribute("id"),'site'),',',lc_tmp
   END IF
   IF NOT cl_null(adzp150_chk_ent_and_site(p_node.getAttribute("id"),'ent')) THEN
      LET lc_tmp = adzp150_chk_ent_and_site(p_node.getAttribute("id"),'ent'),',',lc_tmp
   END IF
   LET lst_token2 = base.StringTokenizer.create(lc_tmp, ',')
   
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_token = ls_token.trim()
      LET ls_token2 = ls_token2.trim()
      LET ls_token2 = lst_token2.nextToken()
      LET ls_value = ls_value, ls_token, ' = ', ls_token2
      IF lst_token.hasMoreTokens() THEN
         LET ls_value = ls_value, ' AND '
      END IF
   END WHILE
   LET ls_value = ls_value, " \","
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #detail_append_wc
   LET ls_value = ""
   LET ls_name = adzp150_create_name(li_master, "detail_append_detail_wc", "<<<")
   LET ls_value = g_properties.getValue(ls_name)
   LET ls_value = ls_value,"   IF NOT cl_null(g_wc2_table",pi_idx USING "<<<",") THEN \n"
                          ,"      LET g_sql = g_sql CLIPPED,\" AND \", g_wc2_table",pi_idx USING "<<<", " CLIPPED\n"
                          ,"   END IF \n"
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #自己的detail_fill_chk
   LET ls_name = adzp150_create_name(pi_idx, "detail_fill_chk", "<<<")
   LET ls_value = "(NOT cl_null(g_wc2_table",li_master USING "<<<",") AND ",
                  "g_wc2_table",li_master USING "<<<",".trim() <> '1=1') OR \n",
                  "      (NOT cl_null(g_wc2_table",pi_idx USING "<<<",") AND ",
                  "g_wc2_table",pi_idx USING "<<<",".trim() <> '1=1') "
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #上層的detail_fill_chk
   LET ls_name = adzp150_create_name(li_master, "detail_fill_chk", "<<<")
   LET ls_value = g_properties.getValue(ls_name)
   IF NOT cl_null(ls_value) THEN
      LET ls_value = ls_value, ' OR \n      '
   END IF
   LET ls_value = ls_value,
                  "(NOT cl_null(g_wc2_table",pi_idx USING "<<<",") AND ",
                  "g_wc2_table",pi_idx USING "<<<",".trim() <> '1=1') "
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #detail_fill_chk_all
   LET ls_name = "detail_fill_chk_all"
   LET ls_value = g_properties.getValue(ls_name)
   IF NOT cl_null(ls_value) THEN
      LET ls_value = ls_value, ' AND \n      '
   END IF
   LET ls_value = ls_value, 
                  "(cl_null(g_wc2_table",pi_idx USING "<<<",") OR ",
                  "g_wc2_table",pi_idx USING "<<<",".trim() = '1=1') "
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #master table
   LET g_detail_table_info[pi_idx].master = g_detail_table_info[li_master].id
   
   #上層單身第一個page名稱(detail_master_first_page_by_tbl)
   LET ls_name = adzp150_create_name(g_detail_table_info[li_master].first_page, "detail_var_title", "<<<")
   LET ls_value = g_properties.getValue(ls_name)
   LET ls_name = adzp150_create_name(pi_idx, "detail_master_first_page_by_tbl", "<<<")
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #上層單身第一個page的key名稱(detail_master_first_key_by_tbl)
   LET ls_value = g_detail_table_info[li_master].first_key
   LET ls_name = adzp150_create_name(pi_idx, "detail_master_first_key_by_tbl", "<<<")
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #上層單身第一個page頁碼(detail_master_first_page_num_by_tbl)
   LET ls_name = adzp150_create_name(pi_idx, "detail_master_first_page_num_by_tbl", "<<<")
   LET ls_value = g_detail_table_info[li_master].first_page
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #找出這個單身的var_fk(等同於一般單身master_var_pk)
   LET g_detail_table_info[pi_idx].var_fk = g_detail_table_info[li_master].var_fk,',',
                                            g_detail_table_info[li_master].var_pk
   
   LET ls_value = g_detail_table_info[pi_idx].var_fk
   LET ls_name = adzp150_create_name(pi_idx, "detail_var_fk_by_tbl", "<<<")
   CALL g_properties.addAttribute(ls_name,ls_value)

   #mapping_fk
   LET g_detail_table_info[pi_idx].mapping_fk = g_detail_table_info[li_master].fk,',',
                                                g_detail_table_info[li_master].pk
   
   #找個這個單身的fk數量
   LET ls_name = adzp150_create_name(pi_idx, "detail_fk_num_by_tbl", "<<<")
   LET ls_value = adzp150_create_name(li_master, "detail_pk_num_by_tbl", "<<<")
   LET ls_value = ls_value, g_properties.getValue("ls_value")
   
   #在上層單身的before row段呼叫_b_fill2(pi_idx)
   LET ls_master_pages = g_detail_table_info[li_master].page
   LET lst_token = base.StringTokenizer.create(ls_master_pages, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_name = adzp150_create_name(ls_token, "detail_page_fill", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_value = ls_value,"CALL ",g_properties.getValue("general_prefix"),
                              "_b_fill2('",pi_idx USING "<<<", "')\n"
      LET ls_name = adzp150_create_name(ls_token, "detail_page_fill", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
   END WHILE
   
   #var_fk_t(上層的fk)
   LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_master].fk , ',')
   LET li_tmp = 1
   LET ls_value = ""
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_tmp = adzp150_create_name(li_tmp, "master_field_pk", "&&")
      LET ls_value = ls_value, "g_",g_properties.getValue(ls_tmp),"_t"
      LET ls_value = ls_value, ','
      LET li_tmp = li_tmp + 1
   END WHILE
   
   #var_fk_t(上層的pk)
   LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_master].pk , ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_tmp = adzp150_create_name(g_detail_table_info[li_master].first_page, "detail_var_title", "<<<")
      LET ls_tmp = g_properties.getValue(ls_tmp)
      LET ls_value = ls_value, ls_tmp, "_t.",ls_token
      IF lst_token.hasMoreTokens() THEN
         LET ls_value = ls_value, ','
      END IF
   END WHILE
    
   LET g_detail_table_info[pi_idx].var_fk_t = ls_value
   
   LET g_detail_table_info[pi_idx].curr_idx = "[g_detail_idx2]."
   
   LET g_detail_table_info[pi_idx].type = "D"
   
   #sql_wc_s
   LET ls_value = ""
   LET ls_tmp = g_detail_table_info[pi_idx].fk
   LET lst_token = base.StringTokenizer.create(ls_tmp , ',')
   LET li_page_token = g_detail_table_info[li_master].first_page
   LET ls_tmp = g_detail_page_info[li_page_token].var_fk_t
   LET lst_token2 = base.StringTokenizer.create(ls_tmp , ',')
   WHILE lst_token2.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_token2 = lst_token2.nextToken()
      LET ls_value = ls_value, ls_token, " = '\", ", ls_token2, ", \"'"
      IF lst_token2.hasMoreTokens() THEN
         LET ls_value = ls_value, " AND " 
      END IF
   END WHILE
   LET ls_name = adzp150_create_name(pi_idx, "detail_repro_wc_s", "<<<") 
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #delete_wc
   LET ls_value = ""
   LET ls_tmp = g_detail_table_info[pi_idx].fk
   LET lst_token = base.StringTokenizer.create(ls_tmp , ',')
   LET li_page_token = g_detail_table_info[li_master].first_page
   LET ls_tmp = g_detail_page_info[li_page_token].var_fk
   LET lst_token2 = base.StringTokenizer.create(ls_tmp , ',')
   WHILE lst_token2.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_token2 = lst_token2.nextToken()
      LET ls_value = ls_value, ls_token, " = ", ls_token2
      IF lst_token2.hasMoreTokens() THEN
         LET ls_value = ls_value, " AND " 
      END IF
   END WHILE
   LET ls_name = adzp150_create_name(pi_idx, "detail_delete_wc", "<<<") 
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   LET li_page_idx = 0
   LET lst_page_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
   WHILE lst_page_token.hasMoreTokens()
      LET li_page_token = lst_page_token.nextToken()
      LET li_page_idx = li_page_idx + 1
      
      #上層單身第一個page名稱(detail_master_first_page)
      LET ls_name = adzp150_create_name(pi_idx, "detail_master_first_page_by_tbl", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_page_token, "detail_master_first_page", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
    
      #上層單身第一個page的key名稱(detail_master_first_key)
      LET ls_name = adzp150_create_name(pi_idx, "detail_master_first_key_by_tbl", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_page_token, "detail_master_first_key", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
	  
      #上層單身第一個page頁碼(detail_master_first_page_num)
      LET ls_name = adzp150_create_name(pi_idx, "detail_master_first_page_num_by_tbl", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_page_token, "detail_master_first_page_num", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_insert_clear 上層單身新增時清除下層單身的資料
      LET ls_name = "detail_insert_clear"
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_tmp = adzp150_create_name(li_page_token, "detail_var_title", "<<<")
      LET ls_tmp = g_properties.getValue(ls_tmp)
      LET ls_value = ls_value, '            CALL ', ls_tmp, '.clear()\n'
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_master_num
      LET ls_name = adzp150_create_name(li_page_token, "detail_master_num", "<<<")
      LET ls_value = adzp150_create_name(li_master, "", "<<<")
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #var_fk
      LET g_detail_page_info[li_page_token].var_fk = g_detail_table_info[pi_idx].var_fk
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_fks", "<<<")
      LET ls_value = g_detail_page_info[li_page_token].var_fk
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #var_fk_t(上層的fk)
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_master].fk , ',')
      LET li_tmp = 1
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(li_tmp, "master_field_pk", "&&")
         LET ls_value = ls_value, "g_",g_properties.getValue(ls_tmp),"_t"
         LET ls_value = ls_value, ','
         LET li_tmp = li_tmp + 1
      END WHILE
      
      #var_fk_t(上層的pk)
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_master].var_pk , ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_value = ls_value, ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ','
         END IF
      END WHILE
      LET g_detail_page_info[li_page_token].var_fk_t = ls_value
      
      #var_pk_t
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[pi_idx].pk , ',')
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(li_page_token, "detail_var_title", "<<<")
         LET ls_tmp = g_properties.getValue(ls_tmp)
         LET ls_value = ls_value, ls_tmp, "_t.",ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, ','
         END IF
      END WHILE 
      LET g_detail_page_info[li_page_token].var_pk_t = ls_value
      
      #var_pk
      LET ls_tmp = p_node.getAttribute("pk")    
      LET ls_value = ""
      LET lst_token = base.StringTokenizer.create(ls_tmp , ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_tmp = adzp150_create_name(li_page_token,"detail_var_title", "<<<")
         LET ls_tmp = g_properties.getValue(ls_tmp)
         LET ls_value = ls_value, ls_tmp, "[g_detail_idx2].", ls_token
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, "," 
         END IF
      END WHILE
      LET g_detail_page_info[li_page_token].var_pk = ls_value
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_pks", "<<<") 
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #sql_wc
      LET ls_value = ""
      LET ls_tmp = g_detail_table_info[pi_idx].fk,',',g_detail_table_info[pi_idx].pk
      LET lst_token = base.StringTokenizer.create(ls_tmp , ',')
      LET ls_tmp = g_detail_page_info[li_page_token].var_fk,',',g_detail_page_info[li_page_token].var_pk
      LET lst_token2 = base.StringTokenizer.create(ls_tmp , ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_token2 = lst_token2.nextToken()
         LET ls_value = ls_value, ls_token, " = ", ls_token2
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, " AND " 
         END IF
      END WHILE
      LET g_detail_page_info[li_page_token].sql_wc = ls_value
      LET ls_name = adzp150_create_name(li_page_token, "detail_page_sql", "<<<") 
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #sql_wc_t
      LET ls_value = ""
      LET ls_tmp = g_detail_table_info[pi_idx].fk,',',g_detail_table_info[pi_idx].pk
      LET lst_token = base.StringTokenizer.create(ls_tmp , ',')
      LET ls_tmp = g_detail_page_info[li_page_token].var_fk_t,',',g_detail_page_info[li_page_token].var_pk_t
      LET lst_token2 = base.StringTokenizer.create(ls_tmp , ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_token2 = lst_token2.nextToken()
         LET ls_value = ls_value, ls_token, " = ", ls_token2
         IF lst_token.hasMoreTokens() THEN
            LET ls_value = ls_value, " AND " 
         END IF
      END WHILE
      LET g_detail_page_info[li_page_token].sql_wc_t = ls_value
      LET ls_name = adzp150_create_name(li_page_token, "detail_page_sql_t", "<<<") 
      CALL g_properties.addAttribute(ls_name,ls_value)

      #general_page_idx
      LET ls_value = "2"
      LET ls_name = adzp150_create_name(li_page_token, "general_page_idx", "<<<") 
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #detail_page_clear_by_tbl(組合)
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_title", "<<<") 
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_clear = ls_clear, "      CALL ", ls_value,".clear()\n"
      
      #detail_page_delete_by_tbl(組合)
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_title", "<<<") 
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_delete = ls_delete, "      CALL ", ls_value,".deleteElement(",ls_value,".getLength())\n"
      
      #detail_first_field
      LET ls_name = "detail_page_field", li_page_token USING "<<<"
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_page_token, "detail_first_field", "<<<") 
      LET lst_token = base.StringTokenizer.create(ls_value, ',')
      LET ls_value = lst_token.nextToken()
     IF ls_value.getIndexOf('(',1) THEN
        LET ls_value = ls_value.subString(1,ls_value.getIndexOf('(',1)-1)
     END IF
      IF li_page_idx > 1 THEN
         LET ls_tmp = g_detail_table_info[pi_idx].pk
         #key要加上額外編號
         IF ls_tmp.getIndexOf(ls_value,1) > 0 THEN
            LET ls_value = ls_value, "_", li_page_token USING "<<<"
         END IF
      END IF
      CALL g_properties.addAttribute(ls_name,ls_value)
    
      #detail_page_reproduce
      LET ls_name = adzp150_create_name(li_page_token, "detail_var_title", "<<<") 
      LET ls_value = g_properties.getValue(ls_name)
      IF NOT cl_null(ls_page_reproduce) THEN
         LET ls_page_reproduce = ls_page_reproduce, "               "
      END IF
      LET ls_page_reproduce = ls_page_reproduce, 
                              "LET ",ls_value,"[li_reproduce_target].* = ",
                              ls_value,"[li_reproduce].*\n"
      
   END WHILE  
   
   #detail_page_reproduce(回填到每一個page)
   LET lst_page_token = base.StringTokenizer.create(p_node.getAttribute("page"), ',')
   WHILE lst_page_token.hasMoreTokens()
      LET li_page_token = lst_page_token.nextToken() 
      LET ls_name = adzp150_create_name(li_page_token, "detail_page_reproduce", "<<<") 
      LET ls_value = ls_page_reproduce
      CALL g_properties.addAttribute(ls_name,ls_value)
   END WHILE
   
   #detail_page_clear_by_tbl
   LET ls_name = adzp150_create_name(pi_idx, "detail_page_clear_by_tbl", "<<<")  
   LET ls_value = ls_clear
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #detail_page_delete_by_tbl
   LET ls_name = adzp150_create_name(pi_idx, "detail_page_delete_by_tbl", "<<<")  
   LET ls_value = ls_delete
   CALL g_properties.addAttribute(ls_name,ls_value)

END FUNCTION

#+ detail_sql_forupd重組
FUNCTION adzp150_detail_sql_forupd(ps_sql,ps_pages,ps_table)
   DEFINE ps_pages       STRING
   DEFINE ps_table       STRING
   DEFINE ps_sql         STRING
   DEFINE ls_sql         STRING
   DEFINE lst_token      base.StringTokenizer
   DEFINE ls_token       STRING
   DEFINE ls_fields      STRING
   DEFINE ls_name        STRING
   DEFINE ls_fields_list STRING
   
   LET lst_token = base.StringTokenizer.create(ps_pages, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken() 
      
      #尋找對應page的欄位
      LET ls_name = "detail_field_loc", ls_token
      LET ls_fields = ls_fields, g_properties.getValue(ls_name), ','
      
   END WHILE
   
   LET lst_token = base.StringTokenizer.create(ls_fields, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken() 
      
      #確定欄位出自該table
      IF cl_getField(ps_table,ls_token) THEN
         LET ls_fields_list = ls_fields_list, ls_token, ','
         CONTINUE WHILE
      END IF
      
      IF ls_token.getIndexOf('_',1) > 0 AND ls_token.getIndexOf('_desc',1) = 0 THEN
         LET ls_token = ls_token.subString(1,ls_token.getIndexOf('_',1)-1)
         IF cl_getField(ps_table,ls_token) THEN
            LET ls_fields_list = ls_fields_list, ls_token, ','
            CONTINUE WHILE
         END IF
      END IF
      
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_token = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
         IF cl_getField(ps_table,ls_token) THEN
            LET ls_fields_list = ls_fields_list, ls_token, ','
            CONTINUE WHILE
         END IF
      END IF
      
      #LET ls_fields_list = ls_fields_list, "'',"
      
   END WHILE
    
   LET ls_fields_list = ls_fields_list.subString(1,ls_fields_list.getLength()-1)
   
   LET ls_sql = "SELECT ", ls_fields_list, ps_sql.subString(ps_sql.getIndexOf('FROM',1)-1,ps_sql.getLength())
   
   RETURN ls_sql
   
END FUNCTION

PUBLIC FUNCTION adzp150_detail_ctrl_make()
   DEFINE li_cnt           INTEGER
   DEFINE li_idx           INTEGER
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_token         STRING
   DEFINE ls_name          STRING
   DEFINE ls_value         STRING
   DEFINE ls_tmp           STRING
   
   LET li_cnt = g_detail_table_info.getLength()

   #進行資料轉換
   FOR li_idx = 1 TO li_cnt

      #detail_tbl_fields_ctrlp
      LET lst_token = base.StringTokenizer.create(g_detail_table_info[li_idx].page,",")
      LET ls_value = ""
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET ls_name = "detail_fields_ctrlp", ls_token
         LET ls_tmp = g_properties.getValue(ls_name)
         LET ls_value = ls_value, ls_tmp
      END WHILE
      LET ls_value = adzp150_detail_Filter(ls_value,TRUE)
      LET ls_name = adzp150_create_name(li_idx, "detail_tbl_fields_ctrlp", "<<<")
      CALL g_properties.addAttribute(ls_name, ls_value)

   END FOR
      
END FUNCTION

#+ 確認此表是否需額外添加ent,site
PUBLIC FUNCTION adzp150_chk_ent_and_site(ps_table,ps_type)
   DEFINE ps_table         STRING
   DEFINE ps_type          STRING
   DEFINE ls_field         STRING
   DEFINE ls_all_field     STRING
   DEFINE ls_return        STRING
   DEFINE li_idx           INTEGER

   LET ls_field = ps_table.subString(1,ps_table.getIndexOf('_',1)-1),ps_type
   
   LET ls_all_field = g_properties.getValue("general_all_field")
   
   IF cl_getfield(ps_table, ls_field) AND #表格中有此欄位(ent,site)
      ls_all_field.getIndexOf(ls_field,1) = 0 THEN #且畫面上無出現(需特規處理)
      LET ls_return = ls_field 
      IF ps_type = 'site' THEN
         #比對site_pk
         FOR li_idx = 1 TO g_detail_table_info.getLength() 
            IF g_detail_table_info[li_idx].id = ps_table THEN
               IF g_detail_table_info[li_idx].site_pk = 'Y' THEN
                  LET ls_return = ls_field 
               ELSE
                  LET ls_return = ''
               END IF
               EXIT FOR
            END IF
         END FOR
      END IF 
   ELSE
      LET ls_return = ''
   END IF
   
   RETURN ls_return
   
END FUNCTION
    
#PGS-02667 (200804-00006)  -S-
#檢核該欄位是否為數字型態(需額外加工)
FUNCTION adzp150_chk_int_filed(ps_field)
   DEFINE ps_field LIKE type_t.chr50
   DEFINE ls_type  LIKE type_t.chr50
   DEFINE li_cnt   LIKE type_t.num5
  
   #僅PostgreSQL才需要
   &ifdef is_pgs
   #依據dzeb_t取型態
   SELECT COUNT(1) INTO li_cnt FROM dzeb_t 
    WHERE dzeb002 = ps_field
      AND dzeb007 LIKE 'int%'
   IF li_cnt > 0 THEN
      #是數字型態
      DISPLAY "判斷欄位:",ps_field,"是數字!"
      RETURN TRUE
   ELSE
      #非數字型態
      DISPLAY "判斷欄位:",ps_field,"非數字!"
      RETURN FALSE
   END IF
   &endif

END FUNCTION
#PGS-02667 (200804-00006)  -E-

