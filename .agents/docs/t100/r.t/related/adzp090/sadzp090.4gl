{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ================================
  Initialize              20170714   Ernestliou 1.新增程式
  Modify                  20170922   02160      1. 整批執行異動優化:
                                                   (1)查詢條件增加兩個CheckBox:
                                                      僅顯示有異常的表格, 預設勾選.
                                                      僅顯示本人簽出的表格, 預設勾選.
                                                   (2)畫面一開啟直接以兩個CheckBox當作查詢條件.
                                                   (3)左右兩側各增加兩個顯示欄位:驚嘆號,表格簽出者
}

&include "../4gl/sadzp090_mcr.inc"
IMPORT OS
IMPORT security

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzp090_cnst.inc"
&include "../4gl/sadzp200_cnst.inc"
&include "../4gl/sadzp220_cnst.inc"

CONSTANT cs_left  STRING = "sr_left"
CONSTANT cs_right STRING = "sr_right"

GLOBALS "../../cfg/top_global.inc"

&include "../4gl/sadzp090_type.inc"

DEFINE m_left_tables_list  DYNAMIC ARRAY OF T_TABLE_LIST
DEFINE m_right_tables_list DYNAMIC ARRAY OF T_TABLE_LIST

DEFINE ms_lang  STRING
DEFINE ms_DGENV STRING
DEFINE ms_module STRING
DEFINE ms_topind STRING
DEFINE ms_search STRING
DEFINE ms_GUID STRING
DEFINE mb_alter_start BOOLEAN
DEFINE ms_user STRING #20170922
DEFINE ms_enterprise  STRING #20170922
DEFINE mb_enable_alm  BOOLEAN #20170922
DEFINE ms_show_err_only     STRING #20170922
DEFINE ms_show_my_check_out STRING #20170922

DEFINE ms_sql_main_form_modules STRING
DEFINE mo_arguments  T_ARGUMENTS

FUNCTION sadzp090_run(p_left_array,p_right_array,p_arguments)
DEFINE 
  p_left_array   DYNAMIC ARRAY OF T_TABLE_LIST,
  p_right_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  p_arguments    T_ARGUMENTS
DEFINE
  ls_columns STRING 
  
  CALL sadzp090_initialize(p_left_array,p_right_array,p_arguments.*) RETURNING p_left_array,p_right_array
  CALL sadzp090_initial_form()
  CALL sadzp090_start(p_left_array,p_right_array) RETURNING ls_columns 
  CALL sadzp090_finalize()

  RETURN ls_columns
  
END FUNCTION 
  
FUNCTION sadzp090_initialize(p_left_array,p_right_array,p_arguments)
DEFINE
  p_left_array   DYNAMIC ARRAY OF T_TABLE_LIST,
  p_right_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  p_arguments  T_ARGUMENTS
DEFINE
  lo_rtn_left_array   DYNAMIC ARRAY OF T_TABLE_LIST,  
  lo_rtn_right_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  ls_erpalm  STRING, #20170922
  ls_user    STRING  #20170922

  LET mo_arguments.* = p_arguments.*
  
  LET lo_rtn_left_array  = p_left_array
  LET lo_rtn_right_array = p_right_array
  
  &ifndef DEBUG
    LET ms_lang       = g_lang
    LET ms_enterprise = g_enterprise
    LET ms_user       = g_account
  &else
    CALL FGL_GETENV("USER") RETURNING ls_user
    LET ms_lang       = cs_default_lang
    LET ms_enterprise = "99"
    LET ms_user       = ls_user
  &endif

  LET ms_DGENV = FGL_GETENV("DGENV")
  LET ms_topind = NVL(FGL_GETENV("TOPIND"),cs_default_topind)
  #20170922 begin
  #是否啟動ALM
  CALL FGL_GETENV("TOPALM") RETURNING ls_erpalm
  LET mb_enable_alm = IIF(NVL(ls_erpalm,"N") == "Y",TRUE,FALSE)
  #20170922 end
  
  CALL sadzp090_initial_combobox_sql()

  IF lo_rtn_left_array.getLength() = 0 THEN
    LET ms_show_err_only = "Y" #20170922
    LET ms_show_my_check_out = "Y" #20170922
    CALL sadzp090_get_left_table_list("*",NULL) RETURNING lo_rtn_left_array #20170922
  END IF

  RETURN lo_rtn_left_array,lo_rtn_right_array
  
END FUNCTION

FUNCTION sadzp090_initial_form()
DEFINE 
  lw_window   ui.Window,
  lf_form     ui.Form,
  lo_combobox ui.ComboBox,
  ls_cfg_path STRING,
  ls_4st_path STRING,
  ls_img_path STRING
  
  OPTIONS INPUT WRAP
  CLOSE WINDOW SCREEN

  &ifndef DEBUG
    OPEN WINDOW w_sadzp090 WITH FORM cl_ap_formpath("ADZ","sadzp090")
  &else
    OPEN WINDOW w_sadzp090 WITH FORM sadzp090_get_form_path("sadzp090")
  &endif
  
  CURRENT WINDOW IS w_sadzp090
  
  &ifndef DEBUG
  CALL cl_ui_wintitle(1) #工具抬頭名稱
  CALL cl_load_4ad_interface(NULL)
  &endif

  CALL sadzp090_set_form_title('sadzp090',ms_lang)
  CALL sadzp090_set_form_style(ms_lang)

  LET lo_combobox = ui.ComboBox.forName("formonly.cb_moduleselect")
  CALL sadzp090_fill_combobox(lo_combobox,ms_sql_main_form_modules)
  
END FUNCTION 


FUNCTION sadzp090_start(p_left_array,p_right_array)
DEFINE 
  p_left_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  p_right_array DYNAMIC ARRAY OF T_TABLE_LIST
DEFINE 
  drag_index, drop_index, i INT,
  drag_source STRING,
  drag_value  T_TABLE_LIST
DEFINE 
  lo_left_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  lo_right_array DYNAMIC ARRAY OF T_TABLE_LIST,
  li_index       INTEGER,
  ls_columns     STRING,
  ls_module      STRING,
  ls_search      STRING,
  ls_command     STRING,
  ls_question    STRING,
  ls_replace_arg STRING,
  ls_altering_id STRING,
  lb_altering    BOOLEAN,
  ls_right_table_list STRING,
  lo_dnd         ui.DragDrop
DEFINE
  ls_return STRING   

  LET lo_left_array.*  = p_left_array.*
  LET lo_right_array.* = p_right_array.*

  LET ls_columns = ""
  
  DIALOG ATTRIBUTE(UNBUFFERED)
    DISPLAY ARRAY m_left_tables_list TO sr_left.*
      BEFORE ROW 
        CALL sadzp090_reload(DIALOG)
        
      ON DRAG_START(lo_dnd)
        LET drag_source = cs_left
        LET drag_index = arr_curr()
        LET drag_value.* = m_left_tables_list[drag_index].*

      ON DRAG_FINISHED(lo_dnd)
        INITIALIZE drag_source TO NULL

      ON DRAG_ENTER(lo_dnd)
        IF drag_source IS NULL THEN
          CALL lo_dnd.setOperation(NULL)
        END IF
        
      ON DROP (lo_dnd)
        IF drag_source == cs_left THEN
          CALL lo_dnd.dropInternal()
        ELSE
          LET drop_index = lo_dnd.getLocationRow()
          CALL DIALOG.insertRow(cs_left, drop_index)
          CALL DIALOG.setCurrentRow(cs_left, drop_index)
          LET m_left_tables_list[drop_index].* = drag_value.*
          CALL DIALOG.deleteRow(cs_right, drag_index)
        END IF
        
    END DISPLAY
    
    DISPLAY ARRAY m_right_tables_list TO sr_right.*
      BEFORE ROW 
        CALL sadzp090_reload(DIALOG)
      
      ON DRAG_START(lo_dnd)
        LET drag_source = cs_right
        LET drag_index = arr_curr()
        LET drag_value.* = m_right_tables_list[drag_index].*
        
      ON DRAG_FINISHED(lo_dnd)
        INITIALIZE drag_source TO NULL

      ON DRAG_ENTER(lo_dnd)
        IF drag_source IS NULL THEN
          CALL lo_dnd.setOperation(NULL)
        END IF
        
      ON DROP(lo_dnd)
        IF drag_source == cs_right THEN
          CALL lo_dnd.dropInternal()
        ELSE
          LET drop_index = lo_dnd.getLocationRow()
          CALL DIALOG.insertRow(cs_right, drop_index)
          CALL DIALOG.setCurrentRow(cs_right, drop_index)
          LET m_right_tables_list[drop_index].* = drag_value.*
          CALL DIALOG.deleteRow(cs_left, drag_index)
        END IF
        
    END DISPLAY

    INPUT ls_module FROM cb_ModuleSelect ATTRIBUTE(WITHOUT DEFAULTS)
      ON CHANGE cb_ModuleSelect
        LET ms_module = ls_module
        CALL DIALOG.setCurrentRow("sr_left",1)
        CLEAR sr_left.*
        LET ls_search = ""
        LET ms_module = ls_module
        
        #搜尋, 所以只傳入ms_search
        # Refresh screen array
        CALL sadzp090_get_right_table_list(m_right_tables_list) RETURNING ls_right_table_list
        CALL sadzp090_decode_tables(ls_right_table_list) RETURNING ls_right_table_list
        CALL sadzp090_get_left_table_list(ms_module,ls_right_table_list) RETURNING lo_left_array
        CALL sadzp090_fill_left_columns_list(lo_left_array)
    END INPUT    

    INPUT ls_search FROM edt_Search ATTRIBUTE(WITHOUT DEFAULTS)
      AFTER INPUT
        LET ms_search = sadzp090_replace_string(ls_search,"'","")
      ON CHANGE edt_Search
        LET ms_search = sadzp090_replace_string(ls_search,"'","")
      ON ACTION ACCEPT
        IF ls_search IS NULL THEN 
          LET ls_module = "*" 
          LET ms_module = ls_module
        ELSE
          LET ls_module = "" 
          LET ms_module = ls_module
        END IF  

        LET ms_search = sadzp090_replace_string(ls_search,"'","")
        CALL DIALOG.setCurrentRow("sr_left",1)
        CLEAR sr_left.*
        #搜尋, 所以只傳入ms_search
        # Refresh screen array
        CALL sadzp090_get_right_table_list(m_right_tables_list) RETURNING ls_right_table_list
        CALL sadzp090_decode_tables(ls_right_table_list) RETURNING ls_right_table_list
        CALL sadzp090_get_left_table_list(NVL(ms_module,ms_search),ls_right_table_list) RETURNING lo_left_array
        CALL sadzp090_fill_left_columns_list(lo_left_array)
        
    END INPUT    

    #20170922 begin
    INPUT ms_show_err_only,ms_show_my_check_out FROM lbl_show_err_only,lbl_show_my_check_out ATTRIBUTE(WITHOUT DEFAULTS)
      ON CHANGE lbl_show_err_only
        GOTO _lbl_show_my_check_out
      ON CHANGE lbl_show_my_check_out
        LABEL _lbl_show_my_check_out:
        CALL DIALOG.setCurrentRow("sr_left",1)
        CLEAR sr_left.*
        LET ms_search = sadzp090_replace_string(ls_search,"'","")
        #搜尋, 所以只傳入ms_search
        # Refresh screen array
        CALL sadzp090_get_right_table_list(m_right_tables_list) RETURNING ls_right_table_list
        CALL sadzp090_decode_tables(ls_right_table_list) RETURNING ls_right_table_list
        CALL sadzp090_get_left_table_list(NVL(ms_module,ms_search),ls_right_table_list) RETURNING lo_left_array
        CALL sadzp090_fill_left_columns_list(lo_left_array)
      
    END INPUT
    #20170922 end

    BEFORE DIALOG
      CALL sadzp090_fill_left_columns_list(lo_left_array)
      CALL sadzp090_fill_right_columns_list(lo_right_array)
      CALL sadzp090_reload(DIALOG)
    
    ON ACTION right_move CALL sadzp090_move_to_left_or_right(DIALOG,cs_left,cs_right)
    ON ACTION left_move  CALL sadzp090_move_to_left_or_right(DIALOG,cs_right,cs_left)
    ON ACTION left_up    CALL sadzp090_move_up_or_down(DIALOG,cs_left,TRUE)
    ON ACTION left_down  CALL sadzp090_move_up_or_down(DIALOG,cs_left,FALSE)
    ON ACTION right_up   CALL sadzp090_move_up_or_down(DIALOG,cs_right,TRUE)
    ON ACTION right_down CALL sadzp090_move_up_or_down(DIALOG,cs_right,FALSE)

    ON ACTION find_data
      IF ls_search IS NULL THEN 
        LET ls_module = "*" 
        LET ms_module = ls_module
      ELSE
        LET ls_module = "" 
        LET ms_module = ls_module
      END IF  
      
      CALL DIALOG.setCurrentRow("sr_left",1)
      CLEAR sr_left.*
      LET ms_search = sadzp090_replace_string(ls_search,"'","")
      #搜尋, 所以只傳入ms_search
      # Refresh screen array
      CALL sadzp090_get_right_table_list(m_right_tables_list) RETURNING ls_right_table_list
      CALL sadzp090_decode_tables(ls_right_table_list) RETURNING ls_right_table_list
      CALL sadzp090_get_left_table_list(NVL(ms_module,ms_search),ls_right_table_list) RETURNING lo_left_array
      CALL sadzp090_fill_left_columns_list(lo_left_array)

    ON ACTION btn_exec
      CALL sadzp090_get_if_any_batch_altering_table_exists() RETURNING lb_altering,ls_altering_id
      IF lb_altering THEN
        LET ls_replace_arg = ls_altering_id,"|"
        CALL sadzp000_msg_alert_box(NULL, "adz-01224", ls_replace_arg, 0) RETURNING ls_question
        IF ls_question = cs_response_yes THEN
          #啟動 Table alter monitor
          LET ls_command = "r.r adzp250 -ID '",ls_altering_id,"'"
          DISPLAY cs_command_tag,ls_command
          RUN ls_command WITHOUT WAITING
        END IF        
        CONTINUE DIALOG
      END IF 
      LET ls_columns = ""
      LET mb_alter_start = FALSE
      IF m_right_tables_list.getLength() <> 0 THEN
        CALL sadzp090_generate_interface(cs_batch_alter_name,m_right_tables_list) RETURNING ms_GUID
        LET mb_alter_start = TRUE
        ERROR cs_message_tag,ms_GUID ATTRIBUTES(BLUE)

        #啟動 Table alter manager
        LET ls_command = "r.r adzp245 -ID '",ms_GUID,"'"
        DISPLAY cs_command_tag,ls_command
        RUN ls_command WITHOUT WAITING

        SLEEP 3
        
        #啟動 Table alter monitor
        LET ls_command = "r.r adzp250 -ID '",ms_GUID,"'"
        DISPLAY cs_command_tag,ls_command
        RUN ls_command WITHOUT WAITING
      ELSE
        CALL sadzp000_msg_show_error('adz-01225', 'adz-01225', '', 1)
      END IF
      
    ON ACTION btn_exit
      LET ls_columns = cs_result_cancel
      EXIT DIALOG
      
    ON ACTION CLOSE
      EXIT DIALOG
  
  END DIALOG

  LET ls_return = ls_columns

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzp090_finalize()
  CLOSE WINDOW w_sadzp090
END FUNCTION

FUNCTION sadzp090_reload(p_dialog)
DEFINE 
  p_dialog ui.Dialog
  
  CALL p_dialog.setActionActive("right_move", m_left_tables_list.getLength()  > 0)
  CALL p_dialog.setActionActive("left_move" , m_right_tables_list.getLength() > 0)
  CALL p_dialog.setActionActive("left_up"   , p_dialog.getCurrentRow(cs_left)  > 1)
  CALL p_dialog.setActionActive("left_down" , p_dialog.getCurrentRow(cs_left)  < m_left_tables_list.getLength())
  CALL p_dialog.setActionActive("right_up"  , p_dialog.getCurrentRow(cs_right) > 1)
  CALL p_dialog.setActionActive("right_down", p_dialog.getCurrentRow(cs_right) < m_right_tables_list.getLength())
  
END FUNCTION

FUNCTION sadzp090_get_value(p_list_name,p_index)
DEFINE 
  p_list_name STRING,
  p_index     INT
  
  IF p_list_name = cs_left THEN
    RETURN m_left_tables_list[p_index].*
  ELSE
    RETURN m_right_tables_list[p_index].* 
  END IF
  
END FUNCTION

FUNCTION sadzp090_set_value(p_list_name,p_index,p_value)
DEFINE 
  p_list_name STRING
DEFINE 
  p_index INT,
  p_value T_TABLE_LIST
  
  IF p_list_name = cs_left THEN 
    LET m_left_tables_list[p_index].* = p_value.*
  ELSE
    LET m_right_tables_list[p_index].* = p_value.* 
  END IF
  
END FUNCTION

-- Remove the current item from the source side and move it the destination side
FUNCTION sadzp090_move_to_left_or_right(p_dialog,p_source,p_dest)
DEFINE 
  p_dialog ui.Dialog,
  p_source STRING,
  p_dest   STRING
DEFINE 
  ls_source   STRING,
  ls_dest     STRING,
  li_idx_src  INT, 
  li_idx_dest INT, 
  lo_col_list T_TABLE_LIST
  
  LET ls_source = p_source 
  LET ls_dest   = p_dest 

  {  
  LET li_idx_src  = p_dialog.getCurrentRow(ls_source)
  LET li_idx_dest = p_dialog.getCurrentRow(ls_dest)

  LET li_idx_dest = IIF(li_idx_dest == 0,1,li_idx_dest)
  
  CALL p_dialog.InsertRow(ls_dest,li_idx_dest)
  }

  CALL p_dialog.appendRow(ls_dest)

  LET li_idx_src  = p_dialog.getCurrentRow(ls_source)
  LET li_idx_dest = p_dialog.getArrayLength(ls_dest) -- getCurrentRow(ls_dest)
  LET li_idx_dest = IIF(li_idx_dest == 0,1,li_idx_dest)
  
  #CALL p_dialog.InsertRow(ls_dest,li_idx_dest)
  
  CALL sadzp090_get_value(ls_source,li_idx_src) RETURNING lo_col_list.*
  
  CALL sadzp090_set_value(ls_dest,li_idx_dest,lo_col_list.*)
  CALL p_dialog.deleteRow(ls_source,li_idx_src)
  CALL p_dialog.setCurrentRow(ls_dest,li_idx_dest)
  CALL sadzp090_reload(p_dialog)
  
END FUNCTION

-- Move the current item at the given side 1 row up or down
FUNCTION sadzp090_move_up_or_down(p_dialog,p_side,p_up)
DEFINE 
  p_dialog ui.Dialog,
  p_side   STRING,
  p_up     INT
DEFINE 
  li_idx  INT,
  li_step INT,
  lo_col_list_step  T_TABLE_LIST,
  lo_col_list       T_TABLE_LIST
  
  IF p_up THEN
    LET li_step = -1
  ELSE
    LET li_step = 1
  END IF
  
  LET li_idx = p_dialog.getCurrentRow(p_side)

  CALL sadzp090_get_value(p_side,li_idx) RETURNING lo_col_list.*
  CALL sadzp090_get_value(p_side,li_idx + li_step) RETURNING lo_col_list_step.*
  
  CALL sadzp090_set_value(p_side,li_idx + li_step,lo_col_list.*)
  CALL sadzp090_set_value(p_side,li_idx,lo_col_list_step.*)
  
  CALL p_dialog.setCurrentRow(p_side,li_idx + li_step)
  CALL sadzp090_reload(p_dialog)
  
END FUNCTION

FUNCTION sadzp090_fill_left_columns_list(p_array)
DEFINE 
  p_array  DYNAMIC ARRAY OF T_TABLE_LIST
DEFINE 
  lo_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  li_index  INTEGER

  LET lo_array.* = p_array.*

  CALL m_left_tables_list.Clear();

  FOR li_index = 1 TO lo_array.getLength()
    LET m_left_tables_list[li_index].RECORDTYPE   = lo_array[li_index].RECORDTYPE
    LET m_left_tables_list[li_index].TABLE_NAME   = lo_array[li_index].TABLE_NAME
    LET m_left_tables_list[li_index].TABLE_DESC   = lo_array[li_index].TABLE_DESC
    LET m_left_tables_list[li_index].MODULE_NAME  = lo_array[li_index].MODULE_NAME
    LET m_left_tables_list[li_index].ALTER_RESULT = lo_array[li_index].ALTER_RESULT
    LET m_left_tables_list[li_index].CHECKOUT_BY  = lo_array[li_index].CHECKOUT_BY
  END FOR 

END FUNCTION

FUNCTION sadzp090_get_right_table_list(p_array)
DEFINE 
  p_array  DYNAMIC ARRAY OF T_TABLE_LIST
DEFINE 
  lo_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  ls_tables STRING,
  li_index  INTEGER
DEFINE
  ls_return STRING
  
  LET lo_array.* = p_array.*
  LET ls_return = ""

  FOR li_index = 1 TO lo_array.getLength()
    LET ls_tables = ls_tables,lo_array[li_index].TABLE_NAME,","
  END FOR

  LET ls_return = ls_tables.subString(1,ls_tables.getLength()-1)
  
  RETURN ls_return 

END FUNCTION

FUNCTION sadzp090_fill_right_columns_list(p_array)
DEFINE 
  p_array  DYNAMIC ARRAY OF T_TABLE_LIST
DEFINE 
  lo_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  li_index  INTEGER

  LET lo_array.* = p_array.*

  CALL m_right_tables_list.Clear();

  FOR li_index = 1 TO lo_array.getLength()
    LET m_right_tables_list[li_index].RECORDTYPE   = lo_array[li_index].RECORDTYPE
    LET m_right_tables_list[li_index].TABLE_NAME   = lo_array[li_index].TABLE_NAME
    LET m_right_tables_list[li_index].TABLE_DESC   = lo_array[li_index].TABLE_DESC
    LET m_right_tables_list[li_index].MODULE_NAME  = lo_array[li_index].MODULE_NAME
    LET m_right_tables_list[li_index].ALTER_RESULT = lo_array[li_index].ALTER_RESULT
    LET m_right_tables_list[li_index].CHECKOUT_BY  = lo_array[li_index].CHECKOUT_BY
  END FOR 

END FUNCTION

FUNCTION sadzp090_get_form_path(p_form_name)
DEFINE
  p_form_name STRING 
DEFINE
  ls_form_name    STRING,
  ls_path         STRING,
  ls_os_separator STRING,
  ls_form_path    STRING

  LET ls_form_name = p_form_name
  
  LET ls_path = os.Path.pwd()
  
  LET ls_os_separator = os.Path.separator()
  LET ls_form_path = ls_path,ls_os_separator,ls_form_name                      

  RETURN ls_form_path
  
END FUNCTION 

FUNCTION sadzp090_set_form_title(p_form,p_lang) 
DEFINE 
  p_form  STRING,
  p_lang  STRING
DEFINE 
  ls_form      STRING,
  ls_lang      STRING,
  lo_window    ui.Window,
  lf_form      ui.Form,
  ls_cfg_path  STRING,
  ls_4st_path  STRING,
  ls_img_path  STRING,
  ls_icon_path STRING
DEFINE 
  ls_sql    STRING,
  ls_title  VARCHAR(1024) 

  LET ls_form = p_form
  LET ls_lang = p_lang
  
  LET ls_sql = "select gzdel003                ",
               "  from gzdel_t                 ",
               " where gzdel001 = '",ls_form,"'",
               "   and gzdel002 = '",ls_lang,"'"
               
  PREPARE lcur_set_form_title FROM ls_sql
  EXECUTE lcur_set_form_title INTO ls_title
  FREE lcur_set_form_title

  LET lo_window = ui.Window.getCurrent()
  CALL lo_window.setText(ls_title CLIPPED) 

  LET ls_img_path  = os.Path.join(os.Path.join(FGL_GETENV("RES"), "img"), "ui")
  LET ls_icon_path = os.Path.join(os.Path.join(ls_img_path, "logo"), "dsc_logo.ico")
  TRY 
    CALL lo_window.setImage(ls_icon_path)
  CATCH
    DISPLAY cs_warning_tag,"Can not set logo icon !"
  END TRY   
  
END FUNCTION

FUNCTION sadzp090_set_form_style(p_lang)
DEFINE
  p_lang STRING
DEFINE
  ls_lang     STRING,
  ls_cfg_path STRING,
  ls_4st_path STRING

  LET ls_lang = p_lang
  
  LET ls_cfg_path = os.Path.join(FGL_GETENV("ERP"), "cfg")
  LET ls_4st_path = os.Path.join(os.Path.join(os.Path.join(ls_cfg_path, "4st"), ls_lang), "designer.4st")
  
  TRY 
    CALL ui.Interface.loadStyles(ls_4st_path)
  CATCH
    DISPLAY cs_warning_tag,"Can not found 'designer.4st' !"
  END TRY   
  
END FUNCTION 

FUNCTION sadzp090_fill_combobox(p_combobox,p_sql)
DEFINE
  p_combobox ui.combobox,
  p_sql      STRING
DEFINE
  ls_sql      STRING,
  li_index    INTEGER,
  li_count    INTEGER,
  lo_combobox RECORD
                combobox_name VARCHAR(50),
                combobox_desc VARCHAR(255)
              END RECORD

  LET li_index = 0
  LET ls_sql = p_sql

  PREPARE lpre_combobox FROM ls_sql
  DECLARE lcur_combobox SCROLL CURSOR FOR lpre_combobox

  CALL p_combobox.clear()

  LET li_count = 1

  FOREACH lcur_combobox INTO lo_combobox.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    CALL p_combobox.addItem(lo_combobox.combobox_name,lo_combobox.combobox_desc)
    LET li_count = li_count + 1
  END FOREACH

  FREE lcur_combobox
  FREE lpre_combobox

END FUNCTION

FUNCTION sadzp090_initial_combobox_sql()
DEFINE
  ls_customize_sql               STRING,
  ls_modules_tool_table_sql      STRING,
  ls_cust_modules_tool_table_sql STRING

  #begin of module select on main window
  IF (ms_DGENV = cs_dgenv_customize) THEN
    #Begin:20191230
    &ifdef is_msv
    LET ls_customize_sql = "union all                                          ",
                           "select zj.gzzj001                     module_name, ",
                           "       zj.gzzj001+'. '+zol.gzzol003   module_desc  ",
                           "  from gzzj_t zj                                   ",
                           "  left outer join gzzol_t zol                      ",
                           "               on zol.gzzol001 = zj.gzzj003        ",
                           "              and zol.gzzol002 = '",ms_lang,"'     ",
                           " where 1=1                                         ",
                           "   and zj.gzzj002 = 'ERP'                          ", 
                           "   and not exists (                                ",
                           "                    select 1                       ",
                           "                      from gzzo_t zo               ",
                           "                     where zo.gzzo001 = zj.gzzj001 ",
                           "                  )                                "
    &else
    #End:20191230
    LET ls_customize_sql = "union all                                          ",
                           "select zj.gzzj001                     module_name, ",
                           "       zj.gzzj001||'. '||zol.gzzol003 module_desc  ",
                           "  from gzzj_t zj                                   ",
                           "  left outer join gzzol_t zol                      ",
                           "               on zol.gzzol001 = zj.gzzj003        ",
                           "              and zol.gzzol002 = '",ms_lang,"'     ",
                           " where 1=1                                         ",
                           "   and zj.gzzj002 = 'ERP'                          ", 
                           "   and not exists (                                ",
                           "                    select 1                       ",
                           "                      from gzzo_t zo               ",
                           "                     where zo.gzzo001 = zj.gzzj001 ",
                           "                  )                                "
    &endif  
    #20191230
  ELSE
    LET ls_customize_sql = ""
  END IF

  #Begin:20191230
  &ifdef is_msv
  LET ms_sql_main_form_modules = "select zo.gzzo001                     module_name,           ",
                                 "       zo.gzzo001+'. '+zol.gzzol003   module_desc            ",
                                 "  from (                                                     ", #20170425 begin
                                 "        select zo.*                                          ",
                                 "          from gzzo_t zo                                     ",       
                                 "         where 1=1                                           ",       
                                 "           and zo.gzzo002 = 'ERP'                            ",       
                                 "           and zo.gzzo007 = 'sd'                             ",
                                 "        union                                                ",
                                 "        select zo.*                                          ",
                                 "          from gzzo_t zo                                     ",       
                                 "         where 1=1                                           ",       
                                 "           and zo.gzzo002 = 'ERP'                            ",       
                                 "           and zo.gzzo007 <> 'sd'                            ",
                                 "           and (                                             ",
                                 "                 (zo.gzzo007 = '",ms_topind,"')              ",
                                 "                 or                                          ",
                                 "                 ('",ms_dgenv,"' = '",cs_dgenv_customize,"') ", 
                                 "               )                                             ", 
                                 "       ) zo                                                  ", #20170425 end
                                 "  left outer join gzzol_t zol                                ",
                                 "               on zol.gzzol001 = zo.gzzo001                  ",
                                 "              and zol.gzzol002 = '",ms_lang,"'               ",
                                 " where 1=1                                                   ",
                                 ls_customize_sql,                                           
                                 "union all                                                    ",
                                 "select '*'  module_name,                                     ",
                                 "       '*'  module_desc                                      ",
                                 "  from dual                                                  ",
                                 " order by 1                                                  "  
  &else
  #End:20191230
  LET ms_sql_main_form_modules = "select zo.gzzo001                     module_name,           ",
                                 "       zo.gzzo001||'. '||zol.gzzol003 module_desc            ",
                                 "  from (                                                     ", #20170425 begin
                                 "        select zo.*                                          ",
                                 "          from gzzo_t zo                                     ",       
                                 "         where 1=1                                           ",       
                                 "           and zo.gzzo002 = 'ERP'                            ",       
                                 "           and zo.gzzo007 = 'sd'                             ",
                                 "        union                                                ",
                                 "        select zo.*                                          ",
                                 "          from gzzo_t zo                                     ",       
                                 "         where 1=1                                           ",       
                                 "           and zo.gzzo002 = 'ERP'                            ",       
                                 "           and zo.gzzo007 <> 'sd'                            ",
                                 "           and (                                             ",
                                 "                 (zo.gzzo007 = '",ms_topind,"')              ",
                                 "                 or                                          ",
                                 "                 ('",ms_dgenv,"' = '",cs_dgenv_customize,"') ", 
                                 "               )                                             ", 
                                 "       ) zo                                                  ", #20170425 end
                                 "  left outer join gzzol_t zol                                ",
                                 "               on zol.gzzol001 = zo.gzzo001                  ",
                                 "              and zol.gzzol002 = '",ms_lang,"'               ",
                                 " where 1=1                                                   ",
                                 ls_customize_sql,                                            
                                 "union all                                                    ",
                                 "select '*'  module_name,                                     ",
                                 "       '*'  module_desc                                      ",
                                 "  from dual                                                  ",
                                 " order by 1                                                  "  
  &endif  
  #20191230
                                 
  #end of module select on main window


END FUNCTION

FUNCTION sadzp090_replace_string(p_string,p_from,p_to)
DEFINE 
  p_string STRING,
  p_from   STRING,
  p_to     STRING
DEFINE 
  ls_string      STRING,
  ls_from        STRING,
  ls_to          STRING,
  lo_string_buf  base.StringBuffer
DEFINE
  ls_return STRING  

  LET ls_string = p_string
  LET ls_from   = p_from
  LET ls_to     = p_to
  
  LET lo_string_buf = base.StringBuffer.create()

  CALL lo_string_buf.clear()
  CALL lo_string_buf.append(ls_string)
  CALL lo_string_buf.replace(ls_from,ls_to,0)
  
  LET ls_return = lo_string_buf.toString()

  RETURN ls_return

END FUNCTION

FUNCTION sadzp090_get_left_table_list(p_search,p_right_list)
DEFINE 
  p_search     STRING,
  p_right_list STRING
DEFINE
  ls_search     STRING,
  ls_right_list STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,         
  ls_sql_cond   STRING,
  ls_sql_alm    STRING, #20170922
  ls_sql_show_err_only STRING, #20170922
  ls_sql_show_my_check_out STRING, #20170922
  ls_sql_right_list STRING,
  lo_table_list DYNAMIC ARRAY OF T_TABLE_LIST
DEFINE  
  lo_return DYNAMIC ARRAY OF T_TABLE_LIST

  LET ls_search = p_search
  LET ls_right_list = p_right_list

  IF (ls_search IS NULL) OR (ls_search = "*") THEN
  
    IF (ls_search IS NULL) THEN 
      LET ls_sql_cond = " and 1=2"
    END IF  
    
    IF (ls_search = "*") THEN 
      LET ls_sql_cond = ""
    END IF
    
  ELSE
    CALL sadzp090_parse_search_contents(ls_search) RETURNING ls_sql_cond
  END IF

  IF ls_right_list IS NOT NULL THEN
    LET ls_sql_right_list = " and ea.dzea001 not in (",ls_right_list,") "
  ELSE
    LET ls_sql_right_list = NULL  
  END IF

  #20170922 begin
  IF mb_enable_alm THEN
    LET ls_sql_alm = " left outer join (                                                 ",
                     "                   select *                                        ",
                     "                     from dzlm_t lmo                               ",
                     "                    where 1=1                                      ",
                     "                      and exists (                                 ",
                     "                                   select 1                        ",
                     "                                     from dzlu_t lu                ",
                     "                                    where lu.dzlu003 = lmo.dzlm012 ",
                     "                                 )                                 ",
                     "                 ) lm on lm.dzlm001 = '",cs_spec_type_table,"'     ",
                     "                     and lm.dzlm002 = ea.dzea001                   ",
                     "                     and lm.dzlm008 = '",cs_check_out,"'           "
    
  ELSE
    LET ls_sql_alm = " left outer join dzlm_t lm on lm.dzlm001 = '",cs_spec_type_table,"' ",
                     "                          and lm.dzlm002 = ea.dzea001               ",
                     "                          and lm.dzlm008 = '",cs_check_out,"'       "
  END IF

  IF ms_show_err_only = "Y" THEN
    LET ls_sql_show_err_only = " and ea.dzea011 <> 'Y' "
  ELSE
    LET ls_sql_show_err_only = ""
  END IF

  IF ms_show_my_check_out = "Y" THEN
    LET ls_sql_show_my_check_out = " and exists (                                         ",
                                   "              select 1                                ",
                                   "                from dzlm_t lmt                       ",
                                   "               where lmt.dzlm001 = 'T'                ",
                                   "                 and lmt.dzlm007 = '",ms_user,"'      ",  
                                   "                 and lmt.dzlm008 = '",cs_check_out,"' ",
                                   "                 and lmt.dzlm002 = ea.dzea001         ",    
                                   "            )                                         "
  ELSE
    LET ls_sql_show_my_check_out = ""
  END IF
  
  #20170922 end  
  
  LET li_rec_cnt = 1
  
  #Begin:20191230
  &ifdef is_msv
  LET ls_sql = "select distinct case ea.dzea011 when 'N' then '!'                    ",
               "                                when 'I' then '#'                    ",
               "                                else '' end           record_type,   ",
               "       ea.dzea001,eal.dzeal003,ea.dzea003,'' alter_result,           ",
               "       lm.dzlm007+'-'+ag.ooag011 checkout_by                         ",
               "  from dzea_t ea                                                     ",
               "  left outer join dzeal_t eal on eal.dzeal001 = ea.dzea001           ",
               "                             and eal.dzeal002 = '",ms_lang,"'        ",
               ls_sql_alm, #20170922
               "  left outer join (                                                  ", #20170922 begin
               "                    select xa.gzxa001,ag.ooag011                     ",
               "                      from ooag_t ag,                                ",
               "                           gzxa_t xa                                 ",
               "                     where xa.gzxaent = ag.ooagent                   ",
               "                       and ag.ooag001 = xa.gzxa003                   ",
               "                       and xa.gzxaent = '",ms_enterprise,"'          ",
               "                     group by xa.gzxa001,ag.ooag011                  ",
               #"                     order by xa.gzxa001                             ",
               "                  ) ag on ag.gzxa001 = lm.dzlm007                    ", #20170922 end
               " where 1=1                                                           ",
               ls_sql_cond,
               ls_sql_right_list,
               ls_sql_show_err_only, #20170922
               ls_sql_show_my_check_out, #20170922
               " order by ea.dzea001                                                 " 
  &else
  #End:20191230
  LET ls_sql = "select distinct case ea.dzea011                               ", 
               "                  when 'N' then '!'                           ",
               "                  when 'I' then '#'                           ",
               "                else '' end  record_type,                     ",
               "       ea.dzea001,eal.dzeal003,ea.dzea003,'' alter_result,    ",
               "       lm.dzlm007||'-'||ag.ooag011 checkout_by                ",
               "  from dzea_t ea                                              ",
               "  left outer join dzeal_t eal on eal.dzeal001 = ea.dzea001    ",
               "                             and eal.dzeal002 = '",ms_lang,"' ",
               ls_sql_alm, #20170922
               "  left outer join (                                           ", #20170922 begin
               "                    select xa.gzxa001,ag.ooag011              ",
               "                      from dsdemo.ooag_t ag,                  ",
               "                           dsdemo.gzxa_t xa                   ",
               "                     where xa.gzxaent = ag.ooagent            ",
               "                       and ag.ooag001 = xa.gzxa003            ",
               "                       and xa.gzxaent = '",ms_enterprise,"'   ",
               "                     group by xa.gzxa001,ag.ooag011           ",
               "                     order by xa.gzxa001                      ",
               "                  ) ag on ag.gzxa001 = lm.dzlm007             ", #20170922 end
               " where 1=1                                                    ",
               ls_sql_cond,
               ls_sql_right_list,
               ls_sql_show_err_only, #20170922
               ls_sql_show_my_check_out, #20170922
               " order by ea.dzea001                                          " 
  &endif
  #20191230
                              
  PREPARE lpre_get_table_list FROM ls_sql
  DECLARE lcur_get_table_list CURSOR FOR lpre_get_table_list

  OPEN lcur_get_table_list
  FOREACH lcur_get_table_list INTO lo_table_list[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_table_list
  CALL lo_table_list.deleteElement(li_rec_cnt)
  
  FREE lpre_get_table_list
  FREE lcur_get_table_list  

  LET lo_return = lo_table_list
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzp090_parse_search_contents(p_search_string)
DEFINE
  p_search_string STRING
DEFINE
  ls_search_string STRING,
  li_loop     INTEGER,
  ls_char     STRING,
  ls_contant  VARCHAR(1024),
  li_curr     INTEGER,
  ls_full_sql STRING,
  lo_search_contents DYNAMIC ARRAY OF T_SEARCH_CONTENTS
DEFINE
  ls_return STRING  

  LET ls_search_string = sadzp090_replace_string(p_search_string,"'","")

  LET ls_contant = ""
  FOR li_loop = 1 TO ls_search_string.getLength()
    LET ls_char = ls_search_string.getCharAt(li_loop) -- .subString(li_loop,li_loop)
    IF ls_char = "," THEN
      CALL sadzp090_trim_str(ls_contant) RETURNING ls_contant
      LET li_curr = lo_search_contents.getLength()+1
      LET lo_search_contents[li_curr].sc_contents = ls_contant
      LET lo_search_contents[li_curr].sc_sql      = sadzp090_generate_contant_sql(ls_contant) 
      LET ls_contant = ""
    ELSE 
      LET ls_contant = ls_contant,ls_char.trim()  
    END IF
  END FOR

  CALL sadzp090_trim_str(ls_contant) RETURNING ls_contant
  IF ls_contant IS NOT NULL THEN 
    LET li_curr = lo_search_contents.getLength()+1
    LET lo_search_contents[li_curr].sc_contents = ls_contant
    LET lo_search_contents[li_curr].sc_sql      = sadzp090_generate_contant_sql(ls_contant) 
    LET ls_contant = ""
  END IF  

  FOR li_loop = 1 TO lo_search_contents.getLength()
    LET ls_full_sql = ls_full_sql,lo_search_contents[li_loop].sc_sql,"\n"
  END FOR 

  LET ls_full_sql = ls_full_sql.subString(ls_full_sql.getIndexOf("or",1)+2,ls_full_sql.getLength())
  
  LET ls_return = " and ( ",NVL(ls_full_sql,"1=1")," )"

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzp090_generate_contant_sql(p_contant)
DEFINE
  p_contant STRING
DEFINE
  ls_content     STRING,
  li_loop        INTEGER,
  ls_char        STRING,
  ls_obj_name    STRING,
  ls_sql         STRING,
  ls_all_sql     STRING,
  ls_contant_sql STRING
DEFINE 
  ls_return  STRING  

  LET ls_content = p_contant

  FOR li_loop = 1 TO ls_content.getLength()
    LET ls_char = ls_content.getCharAt(li_loop) --subString(li_loop,li_loop)
    IF ls_char = "|" THEN
      CALL sadzp090_parse_table_condition(ls_obj_name) RETURNING ls_sql
      LET ls_all_sql = ls_all_sql,ls_sql,"\n"
      LET ls_sql = ""
      LET ls_obj_name = ""
    ELSE 
      LET ls_obj_name = ls_obj_name,ls_char.trim()  
    END IF
  END FOR

  IF ls_obj_name IS NOT NULL THEN 
    CALL sadzp090_parse_table_condition(ls_obj_name) RETURNING ls_sql
    LET ls_all_sql = ls_all_sql,ls_sql,"\n"
    LET ls_sql = ""
    LET ls_obj_name = ""
  END IF    

  LET ls_return = ls_all_sql

  RETURN ls_return  
  
END FUNCTION

FUNCTION sadzp090_parse_table_condition(p_obj_name)
DEFINE 
  p_obj_name STRING 
DEFINE 
  ls_obj_name STRING,
  ls_sql      STRING
DEFINE
  ls_return STRING  

  LET ls_obj_name = p_obj_name.trim()

  IF ls_obj_name.getIndexOf("*",1) > 0 THEN
    CALL sadzp090_replace_string(ls_obj_name,"*","%") RETURNING ls_obj_name
    LET ls_sql = " or ea.dzea001 like '",ls_obj_name,"' "
  ELSE
    LET ls_sql = " or ea.dzea003 = '",ls_obj_name.toUpperCase(),"'      ",
                 " or ea.dzea003 like '%",ls_obj_name.toUpperCase(),"%' ",
                 " or ea.dzea001 like '%",ls_obj_name,"%'               ",
                 " or eal.dzeal003 like '%",ls_obj_name,"%'             "

  END IF

  LET ls_return = ls_sql

  RETURN ls_return
  
END FUNCTION  

FUNCTION sadzp090_trim_str(p_string)
DEFINE
  p_string STRING
DEFINE
  ls_string STRING
DEFINE
  ls_return STRING

  LET ls_string = p_string

  LET ls_string = ls_string.trim()
  
  LET ls_return = ls_string
  
  RETURN ls_return

END FUNCTION

FUNCTION sadzp090_decode_tables(p_tables)
DEFINE
  p_tables STRING
DEFINE
  li_index          INTEGER,
  ls_table_char    STRING,
  ls_tables        STRING,
  ls_tables_string STRING,
  ls_tables_queue  STRING
DEFINE
  ls_return STRING

  LET ls_tables = p_tables

  LET ls_tables_queue  = ""
  LET ls_tables_string = "'"

  FOR li_index = 1 TO ls_tables.getLength()
    LET ls_table_char = NVL(ls_tables.subString(li_index,li_index),cs_null_default)
    IF ls_table_char MATCHES "," THEN
      LET ls_tables_string = ls_tables_string,"'"
      LET ls_tables_queue  = ls_tables_queue,ls_tables_string,","
      LET ls_tables_string = "'"
    ELSE
      LET ls_tables_string = ls_tables_string,ls_table_char
    END IF
  END FOR

  #最後一組或僅有一組要加入Queue
  LET ls_tables_string = ls_tables_string,"'"
  LET ls_tables_queue  = ls_tables_queue,ls_tables_string,","
  LET ls_tables_string = "'"

  LET ls_return = ls_tables_queue.subString(1,ls_tables_queue.getLength()-1)

  IF ls_return = "''" THEN
    LET ls_return = NULL 
  END IF

  RETURN ls_return

END FUNCTION

FUNCTION sadzp090_generate_interface(p_patch_no,p_table_array)
DEFINE
  p_patch_no    STRING,
  p_table_array DYNAMIC ARRAY OF T_TABLE_LIST
DEFINE
  ls_patch_no    STRING,
  lo_table_array DYNAMIC ARRAY OF T_TABLE_LIST,
  lo_dzee_t      T_DZEE_T,
  lo_dzez_t      T_DZEZ_T,
  ls_GUID        STRING,
  li_loop        INTEGER,
  ldt_datetime   DATETIME YEAR TO SECOND
DEFINE
  ls_return STRING  

  LET ls_patch_no = p_patch_no
  LET lo_table_array = p_table_array

  LET ls_GUID = security.RandomGenerator.CreateUUIDString()

  BEGIN WORK 
  
  &ifndef DEBUG
  LET ldt_datetime = cl_get_current()
  &else
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif
  
  LET lo_dzee_t.DZEE001 = ls_GUID
  LET lo_dzee_t.DZEE002 = ci_init_batch_seq_no
  LET lo_dzee_t.DZEE003 = ls_patch_no
  LET lo_dzee_t.DZEE004 = ls_patch_no
  LET lo_dzee_t.DZEE005 = cs_state_processing
  LET lo_dzee_t.DZEE006 = ldt_datetime
  LET lo_dzee_t.DZEE007 = ldt_datetime
  LET lo_dzee_t.DZEE008 = NULL
  LET lo_dzee_t.DZEE009 = NULL
  LET lo_dzee_t.DZEE010 = NULL
  LET lo_dzee_t.DZEE011 = NULL
  LET lo_dzee_t.DZEE013 = cs_db_obj_type_erp_table
  CALL sadzp090_insert_update_dzee_t(lo_dzee_t.*)

  FOR li_loop = 1 TO lo_table_array.getLength()
    LET lo_dzez_t.DZEZ001 = ls_GUID
    LET lo_dzez_t.DZEZ002 = lo_table_array[li_loop].TABLE_NAME
    LET lo_dzez_t.DZEZ003 = cs_state_waiting
    LET lo_dzez_t.DZEZ004 = ls_patch_no
    LET lo_dzez_t.DZEZ005 = NULL
    LET lo_dzez_t.DZEZ006 = NULL
    LET lo_dzez_t.DZEZ007 = lo_table_array[li_loop].MODULE_NAME
    LET lo_dzez_t.DZEZ008 = cs_db_obj_type_erp_table
    #Insert 到DZEZ去
    CALL sadzp090_insert_update_dzez_t(lo_dzez_t.*)
  END FOR
  
  COMMIT WORK

  LET ls_return = ls_GUID
  
  RETURN ls_return

END FUNCTION 

FUNCTION sadzp090_insert_update_dzee_t(p_dzee_t)
DEFINE
  p_dzee_t  T_DZEE_T
DEFINE 
  lo_dzee_t T_DZEE_T

  LET lo_dzee_t.* = p_dzee_t.*

  TRY 
    INSERT INTO dzee_t(
                        dzee001,dzee002,dzee003,dzee004,dzee005,           
                        dzee006,dzee007,dzee008,dzee009,dzee010,           
                        dzee011,dzee013                                    
                      )
               VALUES (
                        lo_dzee_t.dzee001,lo_dzee_t.dzee002,lo_dzee_t.dzee003,lo_dzee_t.dzee004,lo_dzee_t.dzee005,
                        lo_dzee_t.dzee006,lo_dzee_t.dzee007,lo_dzee_t.dzee008,lo_dzee_t.dzee009,lo_dzee_t.dzee010,
                        lo_dzee_t.dzee011,lo_dzee_t.dzee013 
                      )
  CATCH
    DISPLAY cs_warning_tag," Insert DZEE_T unsuccess."
    TRY
      UPDATE dzee_t                                 
         SET dzee003 = lo_dzee_t.dzee003,
             dzee004 = lo_dzee_t.dzee004,
             dzee005 = lo_dzee_t.dzee005,
             dzee006 = coalesce(dzee006,lo_dzee_t.dzee006),
             dzee007 = coalesce(dzee007,lo_dzee_t.dzee007),
             dzee008 = lo_dzee_t.dzee008,
             dzee009 = lo_dzee_t.dzee009,
             dzee010 = lo_dzee_t.dzee010,
             dzee011 = lo_dzee_t.dzee011,
             dzee013 = lo_dzee_t.dzee013 
       WHERE dzee001 = lo_dzee_t.dzee001
         AND dzee002 = lo_dzee_t.dzee002
    CATCH
      DISPLAY cs_error_tag," Insert or Update DZEE_T unsuccess : ",SQLCA.sqlcode
    END TRY    
  END TRY

END FUNCTION

FUNCTION sadzp090_insert_update_dzez_t(p_dzez_t)
DEFINE
  p_dzez_t  T_DZEZ_T
DEFINE 
  lo_dzez_t T_DZEZ_T

  LET lo_dzez_t.* = p_dzez_t.*

  TRY 
    INSERT INTO dzez_t(
                        dzez001,dzez002,dzez003,dzez004,dzez005,
                        dzez006,dzez007,dzez008 
                      )
               VALUES (
                        lo_dzez_t.dzez001,lo_dzez_t.dzez002,lo_dzez_t.dzez003,lo_dzez_t.dzez004,lo_dzez_t.dzez005,
                        lo_dzez_t.dzez006,lo_dzez_t.dzez007,lo_dzez_t.dzez008 
                      )
  CATCH
    DISPLAY cs_warning_tag," Insert DZEZ_T unsuccess."
    TRY
      UPDATE dzez_t                                 
         SET dzez003 = coalesce(dzez003,lo_dzez_t.dzez003),
             dzez004 = coalesce(dzez004,lo_dzez_t.dzez004),
             dzez005 = coalesce(dzez005,lo_dzez_t.dzez005),
             dzez006 = coalesce(dzez006,lo_dzez_t.dzez006),
             dzez007 = coalesce(dzez007,lo_dzez_t.dzez007),
             dzez008 = coalesce(dzez007,lo_dzez_t.dzez008)
       WHERE dzez001 = lo_dzez_t.dzez001
         AND dzez002 = lo_dzez_t.dzez002
    CATCH
      DISPLAY cs_error_tag," Insert or Update DZEZ_T unsuccess : ",SQLCA.sqlcode
    END TRY    
  END TRY

END FUNCTION

FUNCTION sadzp090_get_if_any_batch_altering_table_exists()
DEFINE
  ls_sql         STRING,
  li_rec_count   INTEGER,
  lv_altering_id VARCHAR(60)
DEFINE  
  lb_return BOOLEAN,
  ls_return STRING

  LET ls_sql = "select count(1) cnts,max(dzee001) dzee001  ", 
               "  from dzee_t                              ", 
               " where dzee003 = '",cs_batch_alter_name,"' ",
               "   and dzee005 = '",cs_state_processing,"' " 
                  
  PREPARE lpre_if_any_batch_altering_table_exists FROM ls_sql
  DECLARE lcur_if_any_batch_altering_table_exists CURSOR FOR lpre_if_any_batch_altering_table_exists
  OPEN lcur_if_any_batch_altering_table_exists
  FETCH lcur_if_any_batch_altering_table_exists INTO li_rec_count,lv_altering_id
  CLOSE lcur_if_any_batch_altering_table_exists
  FREE lcur_if_any_batch_altering_table_exists
  FREE lpre_if_any_batch_altering_table_exists

  LET lb_return = IIF(NVL(li_rec_count,0)=0,FALSE,TRUE)
  LET ls_return = lv_altering_id
  
  RETURN lb_return,ls_return
  
END FUNCTION
