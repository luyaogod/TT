# Program name   : sadzp060_8.4gl
# Description    : 本版次調整更新
# Modify         : 20160923 160922-00039 by Hiko : 新建程式.

IMPORT os
IMPORT xml
IMPORT util

SCHEMA ds

GLOBALS "../../cfg/top_global.inc"

PUBLIC FUNCTION sadzp060_8_modify_curr_ver(p_prog, p_ver, p_module, p_kind)
   DEFINE p_prog   LIKE dzaf_t.dzaf001, #程式編號
          p_ver    LIKE dzaf_t.dzaf003, #規格版次(SPEC)/程式版次(CODE)
          p_module LIKE dzaf_t.dzaf006, #目前模組
          p_kind   STRING               #更新種類(SPEC/CODE)
   DEFINE lb_result    BOOLEAN,
          ls_err       STRING,
          ls_file_kind STRING, #SPEC=tsd/CODE=tap
          ls_file_path STRING,
          ldoc_file    xml.DomDocument,
          lnode_file   xml.DomNode 
          
   LET p_prog = p_prog CLIPPED
   LET p_module = UPSHIFT(p_module) CLIPPED
   LET p_kind = UPSHIFT(p_kind) CLIPPED

   DISPLAY "call sadzp060_8() start!"

   IF p_kind="SPEC" THEN
      LET ls_file_kind = "tsd"
   ELSE
      LET ls_file_kind = "tap"
   END IF

   LET ls_file_path = os.path.join(os.path.join(os.path.join(FGL_GETENV(p_module), "dzx"), ls_file_kind), p_prog||"."||ls_file_kind)
   IF NOT os.Path.exists(ls_file_path) THEN
      LET ls_err = "ERROR : ",cl_replace_err_msg(cl_getmsg("adz-00175", g_lang), p_prog||"."||ls_file_kind)
      DISPLAY ls_err
      RETURN FALSE,ls_err
   END IF

   LET ldoc_file = xml.domDocument.create()
   CALL ldoc_file.load(ls_file_path)
   LET lnode_file = ldoc_file.getDocumentElement()

   CALL sadzp060_8_modify_status(lnode_file, p_ver) RETURNING lb_result,ls_err
   IF lb_result THEN
      CALL sadzp060_8_save_file(ldoc_file, ls_file_path) RETURNING lb_result,ls_err
      IF NOT lb_result THEN
         RETURN FALSE,ls_err
      END IF
   ELSE
      RETURN FALSE,ls_err
   END IF

   DISPLAY "call sadzp060_8() finish!"
   RETURN TRUE,NULL
END FUNCTION

PRIVATE FUNCTION sadzp060_8_modify_status(p_node, p_ver)
   DEFINE p_node xml.DomNode,
          p_ver  LIKE dzaf_t.dzaf003
   DEFINE la_base_act  DYNAMIC ARRAY OF RECORD
                       act LIKE dzfu_t.dzfu001
                       END RECORD,
          li_i         SMALLINT,
          obj_base_act util.JSONObject
   DEFINE lnode_child  xml.DomNode,
          lnode_child2 xml.DomNode,
          lnode_child3 xml.DomNode,
          ls_tag_name  STRING,
          ls_err       STRING

   #版次為1的話,基本上就全部更新.

   TRY
      #取得標準ON ACTION的清單.
      LET obj_base_act = util.JSONObject.create()
      DECLARE dzfu_curs CURSOR FOR SELECT DISTINCT(dzfu002) FROM dzfu_t WHERE dzfustus='Y'
      LET li_i = 1
      FOREACH dzfu_curs INTO la_base_act[li_i].act
         CALL obj_base_act.put(la_base_act[li_i].act, li_i)
         LET li_i = li_i + 1
      END FOREACH

      LET lnode_child = p_node.getFirstChildElement()
      WHILE (lnode_child IS NOT NULL)
         LET ls_tag_name = lnode_child.getLocalName()
         CASE
            WHEN ls_tag_name.equals("point") OR 
                 ls_tag_name.equals("section") 
                 IF p_ver=1 THEN
                    #第1版次有程式內容才需要異動.
                    IF sadzp060_8_has_cdata(lnode_child) THEN
                       IF ls_tag_name.equals("section") THEN
                          #section加嚴控管.
                          IF lnode_child.getAttribute("ch")="Y" THEN
                             CALL sadzp060_8_set_status(lnode_child)
                          END IF
                       ELSE
                          CALL sadzp060_8_set_status(lnode_child)
                       END IF
                    END IF
                 ELSE
                    IF lnode_child.getAttribute("ch")="Y" THEN
                       CALL sadzp060_8_set_status(lnode_child)
                    END IF
                 END IF
            WHEN ls_tag_name.equals("all") OR
                 ls_tag_name.equals("mi_all") OR
                 ls_tag_name.equals("db_all") OR
                 ls_tag_name.equals("di_all") OR
                 ls_tag_name.equals("act")
                 IF p_ver=1 THEN
                    #第1版次有規格描述才需要異動.
                    IF sadzp060_8_has_cdata(lnode_child) THEN
                       CALL sadzp060_8_set_status(lnode_child)
                    ELSE
                       #非標準ON ACTION有可能沒有規格.
                       IF ls_tag_name.equals("act") THEN
                          IF NOT obj_base_act.has(lnode_child.getAttribute("id")) THEN
                             CALL sadzp060_8_set_status(lnode_child)
                          END IF
                       END IF
                    END IF
                 ELSE
                    IF lnode_child.getAttribute("ch")="Y" THEN
                       CALL sadzp060_8_set_status(lnode_child)
                    END IF
                 END IF

            WHEN ls_tag_name.equals("tree") OR
                 ls_tag_name.equals("field") 
                 #第1層判斷.
                 IF p_ver=1 OR lnode_child.getAttribute("ch")="Y" THEN
                    CALL sadzp060_8_set_status(lnode_child)
                 END IF
                
            WHEN ls_tag_name.equals("other") OR #不論SPEC或CODE都要完整異動<other>.
                 ls_tag_name.equals("table") OR
                 ls_tag_name.equals("prog_rel") OR
                 ls_tag_name.equals("ref_field") OR
                 ls_tag_name.equals("multi_lang") OR
                 ls_tag_name.equals("help_code") OR
                 ls_tag_name.equals("exclude") OR
                 ls_tag_name.equals("strings") #多語言完整異動.
                 #第2層判斷.
                 LET lnode_child2 = lnode_child.getFirstChildElement()
                 WHILE lnode_child2 IS NOT NULL
                    IF ls_tag_name.equals("strings") THEN
                       CALL lnode_child2.setAttribute("lstr", "u")
                    ELSE
                       IF ls_tag_name.equals("help_code") THEN
                          #註記碼得判斷搜尋欄位(help_find)是否有設定.
                          IF lnode_child2.getAttribute("ch")="Y" AND NOT cl_null(lnode_child2.getAttribute("help_find")) THEN
                             CALL sadzp060_8_set_status(lnode_child2)
                          END IF
                       ELSE
                          IF p_ver=1 OR lnode_child2.getAttribute("ch")="Y" OR ls_tag_name.equals("other") THEN 
                             CALL sadzp060_8_set_status(lnode_child2)
                          END IF
                       
                          #table第3層為<sr>
                          IF ls_tag_name.equals("table") THEN
                             LET lnode_child3 = lnode_child2.getFirstChildElement()
                             WHILE lnode_child3 IS NOT NULL
                                IF p_ver=1 OR lnode_child3.getAttribute("ch")="Y" THEN
                                   CALL sadzp060_8_set_status(lnode_child3)
                                END IF
                          
                                LET lnode_child3 = lnode_child3.getNextSiblingElement()
                             END WHILE
                          END IF
                       END IF #ls_tag_name.equals("help_code")
                    END IF #ls_tag_name.equals("strings")
                 
                    LET lnode_child2 = lnode_child2.getNextSiblingElement()
                 END WHILE
         END CASE

         LET lnode_child = lnode_child.getNextSiblingElement() #取得下一個子節點
      END WHILE

      RETURN TRUE,NULL
   CATCH
      LET ls_err = "ERROR : call sadzp060_8_modify_status() error! STATUS:",STATUS
      DISPLAY ls_err
      RETURN FALSE,ls_err
   END TRY
END FUNCTION

PRIVATE FUNCTION sadzp060_8_set_status(p_node)
   DEFINE p_node xml.DomNode
   DEFINE ls_status STRING

   LET ls_status = p_node.getAttribute("status")
   IF cl_null(ls_status) THEN
      CALL p_node.setAttribute("status", "u")
   ELSE
      IF ls_status<>"d" THEN
         CALL p_node.setAttribute("status", "u")
      END IF
   END IF
END FUNCTION

PRIVATE FUNCTION sadzp060_8_has_cdata(p_node)
   DEFINE p_node xml.DomNode
   DEFINE lnode_text  xml.DomNode,
          lnode_cdata xml.DomNode,
          ls_value    STRING

   LET lnode_text = p_node.getFirstChild() 
   IF lnode_text IS NOT NULL THEN
      LET lnode_cdata = lnode_text.getNextSibling() #第二個子節點才可取得到CDATA的資料.
      IF lnode_cdata IS NOT NULL THEN
         LET ls_value = lnode_cdata.getNodeValue()
         IF NOT cl_null(ls_value) THEN
            RETURN TRUE
         END IF
      END IF   
   END IF

   RETURN FALSE
END FUNCTION

PRIVATE FUNCTION sadzp060_8_save_file(p_doc, p_file_path)
   DEFINE p_doc xml.DomDocument,
          p_file_path STRING
   DEFINE ls_err STRING

   IF os.Path.rename(p_file_path, p_file_path||".bak") THEN
      CALL p_doc.setFeature("format-pretty-print", "TRUE")
      CALL p_doc.save(p_file_path)
      
      #最後變成777, 這樣在運行過程中會比較保險.
      IF NOT os.Path.chrwx(p_file_path, 511) THEN
         LET ls_err = "ERROR : ",p_file_path," change rwx failure!"
         DISPLAY ls_err
         RETURN FALSE,ls_err
      END IF
   ELSE
      LET ls_err = "ERROR : ",p_file_path," rename .bak failure!"
      DISPLAY ls_err
      RETURN FALSE,ls_err
   END IF

   RETURN TRUE,NULL
END FUNCTION
