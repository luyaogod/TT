#adzp150 副程式 - 子樣板存取

IMPORT os
SCHEMA ds

GLOBALS "../4gl/adzp150_global.inc"   #180108-00012
GLOBALS
   DEFINE g_properties   om.SaxAttributes
END GLOBALS

#+ 讀取片段樣板並進行取代(欄位檢查段落)
PUBLIC FUNCTION adzp150_make_slice(ps_mdl)
   DEFINE ps_mdl          STRING
   DEFINE ls_mdl          STRING
   DEFINE ls_read         STRING
   DEFINE ls_return       STRING
   DEFINE ls_return_tmp   STRING
   DEFINE ls_text         STRING
   DEFINE lchannel_read   base.Channel
   DEFINE lc_write        base.Channel #回傳用
   DEFINE ls_mdlpath      STRING
   
   #先讀取T100TEMPLATEPATH, 若無則採用$ERP/mdl
   LET ls_mdlpath = FGL_GETENV("T100TEMPLATEPATH")
   IF cl_null(ls_mdlpath) THEN
      LET ls_mdlpath = FGL_GETENV("ERP")
      LET ls_mdlpath = os.Path.join(ls_mdlpath,"mdl")
   END IF
   LET ls_mdlpath = os.Path.join(ls_mdlpath,"slice")
	  
   #定義取用樣板檔案
   LET ls_mdl = "code_",ps_mdl,".template"
   LET ls_mdl = os.Path.join(ls_mdlpath,ls_mdl)
   #開啟樣板檔
   LET lchannel_read = base.Channel.create()
   CALL lchannel_read.setDelimiter("")
   CALL lchannel_read.openFile( ls_mdl CLIPPED, "r" )
   
   WHILE NOT lchannel_read.isEof()
   
      LET ls_text = ""
      LET ls_read = lchannel_read.readLine()
      
      LET ls_return_tmp = ""
      
      #產生code部分
      CASE
         #page段落產生
         WHEN ls_read.getIndexOf("#pages - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,'',0) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #page段落產生(input段落)
         WHEN ls_read.getIndexOf("#pages_input - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,'',1) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""         
            
         #page段落產生(input段落,單身的單身)
         WHEN ls_read.getIndexOf("#pages_input_d - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,'',2) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
         
         #page段落產生(display段落)
         WHEN ls_read.getIndexOf("#pages_display - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,'',3) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""         
            
         #page段落產生(display段落,單身的單身)
         WHEN ls_read.getIndexOf("#pages_display_d - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,'',4) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #page段落產生(單身的單頭)
         WHEN ls_read.getIndexOf("#pages_m - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,'',5) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #page段落產生(單身的單身)
         WHEN ls_read.getIndexOf("#pages_d - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,'',6) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #單頭key段落產生    
         WHEN ls_read.getIndexOf("#master_keys - Start -",1) > 0
            CALL adzp150_make_keys(lchannel_read,'','m') RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #單身key段落產生    
         WHEN ls_read.getIndexOf("#detail_keys - Start -",1) > 0
            CALL adzp150_make_keys(lchannel_read,'','d') RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #多table段落產生    
         WHEN ls_read.getIndexOf("#tables - Start -",1) > 0
            CALL adzp150_make_tables(lchannel_read,'',0) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #多table段落產生(主要table)   
         WHEN ls_read.getIndexOf("#tables_m - Start -",1) > 0
            CALL adzp150_make_tables(lchannel_read,'',1) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #多table段落產生(附帶table)
         WHEN ls_read.getIndexOf("#tables_d - Start -",1) > 0
            CALL adzp150_make_tables(lchannel_read,'',2) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
            
         #多table段落產生(附帶table的table)
         WHEN ls_read.getIndexOf("#tables_d2 - Start -",1) > 0
            CALL adzp150_make_tables(lchannel_read,'',3) RETURNING lchannel_read,lc_write,ls_return_tmp
            LET ls_read = ""
      
         #舊段落,暫不使用
         #page段落產生
         #WHEN ls_read.getIndexOf("#pages - Start -",1) > 0
         #   CALL adzp150_slice_pages(lchannel_read) RETURNING lchannel_read,ls_return_tmp
         #   LET ls_read = ""
         #   
         ##單頭key段落產生
         #WHEN ls_read.getIndexOf("#master_keys - Start -",1) > 0
         #   CALL adzp150_slice_keys(lchannel_read,'m') RETURNING lchannel_read,ls_return_tmp
         #   LET ls_read = ""
         #   
         ##單身key段落產生
         #WHEN ls_read.getIndexOf("#detail_keys - Start -",1) > 0
         #   CALL adzp150_slice_keys(lchannel_read,'d') RETURNING lchannel_read,ls_return_tmp
         #   LET ls_read = ""
         #   
         ##多table段落產生
         #WHEN ls_read.getIndexOf("#tables - Start -",1) > 0
         #   CALL adzp150_slice_tables(lchannel_read) RETURNING lchannel_read,ls_return_tmp
         #   LET ls_read = ""
         
         #slice專用迴圈1
         WHEN ls_read.getIndexOf("#mdls - Start -",1)
            CALL adzp150_make_mdls(lchannel_read,1) RETURNING lchannel_read,ls_return_tmp
            LET ls_read = ""
         
         #slice專用迴圈2
         WHEN ls_read.getIndexOf("#mdls2 - Start -",1)
            CALL adzp150_make_mdls(lchannel_read,2) RETURNING lchannel_read,ls_return_tmp
            LET ls_read = ""         
         
         #slice專用迴圈3
         WHEN ls_read.getIndexOf("#mdls3 - Start -",1)
            CALL adzp150_make_mdls(lchannel_read,3) RETURNING lchannel_read,ls_return_tmp
            LET ls_read = ""
            
      END CASE
      
      LET ls_return = ls_return, ls_return_tmp

      #行代換/ 對 ${} 置換
      IF ls_read.getIndexOf("${",1) AND 
         ( ls_read.getIndexOf("${",1) < ls_read.getIndexOf("}",1) ) THEN
         LET ls_text = adzp150_line_replace(ls_read)
      ELSE
         LET ls_text = ls_read
      END IF
      LET ls_return = ls_return, ls_text, '\n'
   END WHILE  
   
   CALL lchannel_read.close()

   RETURN ls_return

END FUNCTION


##+ 根據tabx設定之page數量產生對應之page段落
#PRIVATE FUNCTION adzp150_slice_pages(l_read)
#   DEFINE l_read        base.Channel
#   DEFINE l_tmp         STRING 
#   DEFINE l_page        DYNAMIC ARRAY OF STRING 
#   DEFINE l_page_num    LIKE type_t.num10 
#   DEFINE li_cnt        LIKE type_t.num10 
#   DEFINE l_is          STRING 
#   DEFINE ls_return     STRING
#
#   #先取出page段落之樣本
#   WHILE TRUE 
#      LET l_tmp = l_read.readLine()
#      CASE 
#         WHEN l_tmp.getIndexOf("#pages -  End  -",1)
#            EXIT WHILE 
#         WHEN l_tmp.getIndexOf("#master_keys - Start -",1)
#            CALL adzp150_make_keys_in_page(l_read,'m') RETURNING l_tmp,l_read
#         WHEN l_tmp.getIndexOf("#detail_keys - Start -",1)
#            CALL adzp150_make_keys_in_page(l_read,'d') RETURNING l_tmp,l_read
#      END CASE
#      LET l_page[1] = l_page[1],l_tmp , "\n"
#   END WHILE 
#   
#   #取得page總數量
#   LET l_page_num = g_properties.getValue("page")
#   
#   #根據page數量產生對應的page段落並進行資料取代
#   FOR li_cnt = 2 TO l_page_num
#      
#      #先將${page}取代掉
#      LET l_is = li_cnt
#      LET l_page[li_cnt] = adzp150_replace_page(l_page[1],li_cnt)
#                        
#      #針對不同的page進行key取代
#      LET ls_return = ls_return, adzp150_make_keys_in_multi_table(l_page[li_cnt],li_cnt)
#      
#      #若資料未取代完成則重複返回
#      WHILE ls_return.getIndexOf("${",1) > 0 
#         LET ls_return = adzp150_line_replace(ls_return)
#      END WHILE  
#
#   END FOR    
#   
#   RETURN l_read, ls_return
#   
#END FUNCTION 
#
##+ 根據tabx設定之key數量產生對應之key段落
#PRIVATE FUNCTION adzp150_slice_keys(l_read,l_type)
#   DEFINE l_read        base.Channel
#   DEFINE l_type        STRING 
#   DEFINE l_tmp         STRING 
#   DEFINE l_key         DYNAMIC ARRAY OF STRING 
#   DEFINE l_key_num     LIKE type_t.num10 
#   DEFINE li_cnt        LIKE type_t.num10 
#   DEFINE l_is          STRING 
#   DEFINE ls_return     STRING
#   
#   #先取出key段落之樣本
#   WHILE TRUE 
#      LET l_tmp = l_read.readLine()
#      IF l_tmp.getIndexOf("#keys -  End  -",1) THEN
#          EXIT WHILE 
#      END IF
#      LET l_key[1] = l_key[1],l_tmp , "\n"
#   END WHILE 
#
#   #取得key總數量
#   IF l_type = 'm' THEN
#      LET l_key_num = g_properties.getValue("master_pk_num")
#   ELSE
#      LET l_key_num = g_properties.getValue("detail_pk_num")
#   END IF
#
#   LET ls_return = ""
#   
#   #根據key數量產生對應的key段落並進行資料取代
#   FOR li_cnt = 2 TO l_key_num
#      
#      #先將${key}取代掉
#      LET l_is = li_cnt
#      LET ls_return = ls_return, 
#                        adzp150_replace_key(l_key[1],li_cnt)
#                        
#      #若資料未取代完成則重複返回
#      WHILE ls_return.getIndexOf("${",1) > 0 
#         LET ls_return = adzp150_line_replace(ls_return)
#      END WHILE  
#      
#   END FOR   
#   
#   RETURN l_read, ls_return
#   
#END FUNCTION 
#
##+ 根據tabx設定之table數量產生對應之table段落
#PRIVATE FUNCTION adzp150_slice_tables(l_read)
#   DEFINE l_read        base.Channel
#   DEFINE l_tmp         STRING 
#   DEFINE l_table       DYNAMIC ARRAY OF STRING 
#   DEFINE l_table_num   LIKE type_t.num10 
#   DEFINE li_cnt        LIKE type_t.num10 
#   DEFINE ls_return     STRING
#   
#   #先取出table段落之樣本
#   WHILE TRUE 
#      LET l_tmp = l_read.readLine()
#      CASE 
#         WHEN l_tmp.getIndexOf("#tables -  End  -",1)
#            EXIT WHILE 
#         WHEN l_tmp.getIndexOf("#master_keys - Start -",1)
#            CALL adzp150_make_keys_in_table(l_read,'m') RETURNING l_tmp,l_read
#         WHEN l_tmp.getIndexOf("#detail_keys - Start -",1)
#            CALL adzp150_make_keys_in_table(l_read,'d') RETURNING l_tmp,l_read
#      END CASE
#      LET l_table[1] = l_table[1],l_tmp , "\n"
#   END WHILE 
#
#   #取得table總數量
#   LET l_table_num = g_properties.getValue("detail_table_num")
#
#   #根據key數量產生對應的key段落並進行資料取代
#   FOR li_cnt = 2 TO l_table_num
#      
#      #先將${table}取代掉
#      LET ls_return = ls_return, adzp150_replace_table(l_table[1],li_cnt)
#                        
#      #若資料未取代完成則重複返回
#      WHILE ls_return.getIndexOf("${",1) > 0 
#         LET ls_return = adzp150_line_replace(ls_return)
#      END WHILE  
#
#   END FOR    
#   
#   RETURN l_read, ls_return
#   
#END FUNCTION 

#mdl產生
PUBLIC FUNCTION adzp150_make_mdls(ps_read,pi_num)
   DEFINE ps_read             base.Channel
   DEFINE pi_num              LIKE type_t.num5 
   DEFINE ls_mdl              DYNAMIC ARRAY OF STRING #第幾組mdl
   DEFINE li_mdl_num          LIKE type_t.num5 
   DEFINE li_cnt              LIKE type_t.num5 
   DEFINE ls_is               STRING 
   DEFINE ls_return           STRING
   DEFINE ls_tmp              STRING
   
   #先取出mdl段落之樣版
   WHILE TRUE 
      LET ls_tmp = ps_read.readLine()
      IF ls_tmp.getIndexOf("#mdls -  End  -",1) THEN
         LET ls_mdl[1] = ls_mdl[1].subString(1,ls_mdl[1].getLength()-1)
         EXIT WHILE 
      END IF
      LET ls_mdl[1] = ls_mdl[1], ls_tmp, '\n'
   END WHILE 
   
   #總數量
   LET ls_tmp = adzp150_create_name(pi_num, "mdl_mdls", "<<<") 
   LET li_mdl_num = g_properties.getValue(ls_tmp)
   
   #根據mdl數量產生對應的mdl段落並進行資料取代
   FOR li_cnt = 2 TO li_mdl_num
      
      #先將${mdls}取代掉
      LET ls_is = li_cnt
      LET ls_mdl[li_cnt] = adzp150_replace_mdls(ls_mdl[1],li_cnt)
                        
      #若資料未取代完成則重複返回
      WHILE ls_mdl[li_cnt].getIndexOf("${",1) > 0 
         LET ls_mdl[li_cnt] = adzp150_line_replace(ls_mdl[li_cnt])
      END WHILE  
      LET ls_return = ls_return, ls_mdl[li_cnt]
      
      IF li_cnt < li_mdl_num THEN
         LET ls_return = ls_return, '\n'
      END IF
      
   END FOR    
   
   RETURN ps_read, ls_return
   
END FUNCTION 

#將${mdl}取代為號碼
PUBLIC FUNCTION adzp150_replace_mdls(ps_mdl, pi_i)
   DEFINE ps_mdl   STRING 
   DEFINE pi_i     STRING 
   DEFINE li_s     LIKE type_t.num5 
   DEFINE li_e     LIKE type_t.num5 
   DEFINE ls_mdl   STRING 

   WHILE TRUE 
      LET li_s = ps_mdl.getIndexOf("${mdl}",1) - 1
      LET li_e = ps_mdl.getLength()
      #取代${mdl}
      LET ls_mdl = ps_mdl.subString(1,li_s), pi_i, ps_mdl.subString((li_s+7),li_e)
      LET ps_mdl = ls_mdl
      IF ls_mdl.getIndexOf("${mdl}",1) < 1 THEN 
         EXIT WHILE
      END IF 
   END WHILE 
    
   RETURN ls_mdl
   
END FUNCTION 

