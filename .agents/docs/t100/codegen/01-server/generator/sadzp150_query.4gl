#adzp150 副程式 - Q類樣板特別處理&MD5相關行為

IMPORT os
SCHEMA ds

GLOBALS "../4gl/adzp150_global.inc"   #180108-00012
GLOBALS
   DEFINE g_properties   om.SaxAttributes
END GLOBALS

DEFINE g_dzbb RECORD 
         prog_name    STRING,
         point_name   STRING,
         point_ver    STRING,
         addpoint     STRING
         END RECORD

#+ 生成所需的Q類資訊(入口)
PUBLIC FUNCTION adzp150_create_query_entrance(ps_idx)
   DEFINE ps_idx                STRING
   
   CALL adzp150_fields_qbe(ps_idx)
   
END FUNCTION

PUBLIC FUNCTION adzp150_fields_qbe(ps_idx)
   DEFINE ps_idx                 STRING
   DEFINE ls_field_list          STRING
   DEFINE lst_token              base.StringTokenizer
   DEFINE ls_token               STRING
   DEFINE lst_token2             base.StringTokenizer
   DEFINE ls_token2              STRING
   DEFINE ls_name                STRING
   DEFINE ls_value               STRING
   DEFINE ls_object              STRING
   DEFINE ls_field               STRING
   DEFINE ls_display_condition   STRING
   DEFINE ls_display_info        STRING
   DEFINE ls_srfield_all         STRING
   DEFINE ls_fields_qbe          STRING
   DEFINE ls_pages               STRING
   DEFINE ls_idx                 STRING
   DEFINE ls_sr                  STRING
   
   #確定這個table所屬的page
   LET ls_name = adzp150_create_name(ps_idx, "detail_tbl_pages", "<<<")
   LET ls_pages = g_properties.getValue(ls_name)
   LET lst_token = base.StringTokenizer.create(ls_pages, ",")
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      
      LET ls_name = adzp150_create_name(ls_token, "sr_name", "<<<")
      LET ls_sr = g_properties.getValue(ls_name)

      LET ls_name = "detail_fields_qbe", ls_token
      LET ls_field_list = g_properties.getValue(ls_name)
      
      LET lst_token2 = base.StringTokenizer.create(ls_field_list, ",")
      WHILE lst_token2.hasMoreTokens()
         LET ls_token2 = lst_token2.nextToken()
         LET ls_object = ls_token2.subString(ls_token2.getIndexOf('(',1)+1,ls_token2.getLength()-1)
         LET ls_field = ls_token2.subString(1,ls_token2.getIndexOf('(',1)-1)
         LET ls_display_condition = ls_display_condition, 
                                    3 SPACES, "CALL ${general_prefix}_filter_show('",ls_field,"','",ls_object,"')", "\n"
         LET ls_display_info = ls_display_info, 
                               12 SPACES, "DISPLAY ${general_prefix}_filter_parser('",ls_field,"')",
                               " TO ", ls_sr, "[1].", ls_object, "\n"
         LET ls_srfield_all = ls_srfield_all, ls_sr, "[1].", ls_object, ","
         LET ls_fields_qbe = ls_fields_qbe, ls_field,","
         
      END WHILE

   END WHILE
   LET ls_srfield_all = ls_srfield_all.subString(1,ls_srfield_all.getLength()-1)
   LET ls_fields_qbe = ls_fields_qbe.subString(1,ls_fields_qbe.getLength()-1)
   
   #detail_tbl_srfield_all(tbl)
   LET ls_name = adzp150_create_name(ps_idx, "detail_tbl_srfield_all", "<<<")
   LET ls_value = ls_srfield_all
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #detail_display_info(tbl)
   LET ls_name = adzp150_create_name(ps_idx, "detail_display_info", "<<<")
   LET ls_value = ls_display_info
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #detail_tbl_fields_qbe(tbl)
   LET ls_name = adzp150_create_name(ps_idx, "detail_tbl_fields_qbe", "<<<")
   LET ls_value = ls_fields_qbe
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #detail_display_condition(一個)
   LET ls_name = "detail_display_condition"
   LET ls_value = g_properties.getValue("detail_display_condition"), ls_display_condition
   CALL g_properties.addAttribute(ls_name,ls_value)
   
END FUNCTION

#單頭用
PUBLIC FUNCTION adzp150_fields_qbe_filter(ps_fields)
   DEFINE ps_fields              STRING
   DEFINE ls_fields              STRING
   DEFINE ls_tmp                 STRING
   DEFINE lst_token              base.StringTokenizer
   DEFINE ls_token               STRING
   
   LET lst_token = base.StringTokenizer.create(ps_fields, ",")
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_tmp = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
      ELSE
         LET ls_tmp = ls_token
      END IF
      LET ls_fields = ls_fields, ls_tmp
      IF lst_token.hasMoreTokens() THEN
         LET ls_fields = ls_fields, ','
      END IF
   END WHILE
   
   RETURN ls_fields
   
END FUNCTION

#單頭用
PUBLIC FUNCTION adzp150_browser_filter_define(ps_page,ps_fields)
   DEFINE ps_fields               STRING
   DEFINE ps_page                 STRING
   DEFINE ls_fields               STRING
   DEFINE ls_sr_name              STRING
   DEFINE ls_name                 STRING
   DEFINE ls_value                STRING
   DEFINE ls_object               STRING
   DEFINE ls_field                STRING
   DEFINE lst_token               base.StringTokenizer
   DEFINE ls_token                STRING
   DEFINE ls_fields_filter        STRING
   DEFINE ls_display_condition    STRING
   DEFINE ls_display_info         STRING
   DEFINE ls_fields_filter_bs     STRING
   
   LET ls_fields_filter     = g_properties.getValue("browser_fields_filter_q")
   LET ls_display_condition = g_properties.getValue("browser_fields_filter_bs_q")
   LET ls_fields_filter_bs  = g_properties.getValue("browser_display_condition_q")
   IF NOT cl_null(ls_fields_filter) THEN
      LET ls_fields_filter = ls_fields_filter, ','
   END IF
   IF NOT cl_null(ls_fields_filter_bs) THEN
      LET ls_fields_filter_bs = ls_fields_filter_bs, ','
   END IF
   
   LET ls_name = adzp150_create_name(ps_page, "browser_sr_name", "<<<")
   LET ls_sr_name = g_properties.getValue(ls_name)
   
   LET lst_token = base.StringTokenizer.create(ps_fields, ",")
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_object = ls_token.subString(ls_token.getIndexOf('(',1)+1,ls_token.getIndexOf(')',1)-1)
      LET ls_field  = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
      LET ls_fields_filter = ls_fields_filter, ls_field
      LET ls_display_condition = ls_display_condition, 
                                 12 SPACES, "DISPLAY ${general_prefix}_filter_parser('",ls_field,"')",
                                 " TO ",ls_sr_name,"[1].", ls_object, "\n"
      LET ls_fields_filter_bs = ls_fields_filter_bs, ls_sr_name,'[1].', ls_object
      LET ls_display_info = ls_display_info,
                            3 SPACES, "CALL ${general_prefix}_filter_show('",ls_field,"','",ls_object,"')", "\n"
      IF lst_token.hasMoreTokens() THEN
         LET ls_fields_filter = ls_fields_filter, ','
         LET ls_fields_filter_bs = ls_fields_filter_bs, ','
      END IF
      
   END WHILE
  
   #browser_fields_filter_q
   LET ls_name = "browser_fields_filter_q"
   LET ls_value = ls_fields_filter
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #browser_fields_filter_bs_q
   LET ls_name = "browser_fields_filter_bs_q"
   LET ls_value = ls_fields_filter_bs
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #browser_display_condition_q
   LET ls_name = "browser_display_condition_q"
   LET ls_value = ls_display_condition 
   CALL g_properties.addAttribute(ls_name,ls_value)
     
   #browser_display_info_q
   LET ls_name = "browser_display_info_q"
   LET ls_value = ls_display_info
   CALL g_properties.addAttribute(ls_name,ls_value) 
   
END FUNCTION
