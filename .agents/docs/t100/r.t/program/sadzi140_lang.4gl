&include "../4gl/sadzi140_mcr.inc"

IMPORT os
IMPORT util

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"

&include "../4gl/sadzp000_type.inc"
&include "../4gl/sadzi140_type.inc"

FUNCTION sadzi140_lang_refresh_column_description(p_table_column,p_lang_type,p_description)
DEFINE
  p_table_column STRING,
  p_lang_type    STRING,
  p_description  STRING
CONSTANT
  lci_max_column_desc_length INTEGER = 490
DEFINE
  ls_table_column STRING,
  ls_lang_type    STRING,
  ls_description  STRING,
  ls_column       STRING,
  lb_result       BOOLEAN  
DEFINE
  lb_return BOOLEAN  

  LET ls_table_column = p_table_column
  LET ls_lang_type    = p_lang_type
  LET ls_description  = p_description.subString(1,IIF(p_description.getLength() >= lci_max_column_desc_length,lci_max_column_desc_length,p_description.getLength()))

  IF ls_description.getLength() = lci_max_column_desc_length THEN
    LET ls_description = ls_description,"..." 
  END IF 

  LET lb_result = TRUE
  
  LET ls_column = ls_table_column.subString(ls_table_column.getIndexOf(".",1)+1,ls_table_column.getLength())
  
  IF (ls_table_column.getIndexOf(".",1) = 0) OR ls_column IS NULL THEN
    DISPLAY cs_error_tag,"Column format not match !! Update column description failed !!"
    LET lb_result = FALSE
  ELSE
    CALL sadzi140_lang_update_column_description(ls_column.toLowerCase(),ls_lang_type,ls_description) RETURNING lb_result
    DISPLAY cs_success_tag,"Update column description successed !!"
  END IF   
  
  LET lb_return = lb_result
  
  RETURN lb_return
  
END FUNCTION 

FUNCTION sadzi140_lang_update_column_description(p_column,p_lang_type,p_description)
DEFINE
  p_column       STRING,
  p_lang_type    STRING,
  p_description  STRING
DEFINE
  ls_column       STRING,
  ls_lang_type    STRING,
  ls_description  VARCHAR(4000),
  ls_update_sql   STRING,
  lb_result       BOOLEAN 
DEFINE 
  lb_return  BOOLEAN 

  LET ls_column      = p_column
  LET ls_lang_type   = p_lang_type
  LET ls_description = p_description

  LET lb_result = TRUE
  
  LET ls_update_sql = "update dzebl_t                       ",
                      "   set dzebl004 = ?                  ",
                      " where dzebl001 = '",ls_column,"'    ", 
                      "   and dzebl002 = '",ls_lang_type,"' ",
                      "   and dzebl004 is null              "
                     
  TRY    
    PREPARE lpre_update_dzebl_t FROM ls_update_sql
    EXECUTE lpre_update_dzebl_t USING ls_description
  CATCH
    LET lb_result = FALSE
    DISPLAY cs_warning_tag,"Update description with column ",ls_column," unsuccess."
  END TRY

  LET lb_return = lb_result

  RETURN lb_return

END FUNCTION
