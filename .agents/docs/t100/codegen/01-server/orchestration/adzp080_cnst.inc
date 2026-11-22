# Modify 230109-00005#1 : 20230110 by 02160 : 新增支援 XR 報表

CONSTANT cs_copy_from      STRING = "[FROM]"
CONSTANT cs_copy_to        STRING = "[ TO ]"

CONSTANT cs_new_line STRING = "\n"

CONSTANT cs_default_lang  STRING = "zh_TW"
CONSTANT cs_erp_path      STRING = "ERP"

#Document Type Define

CONSTANT cs_doc_type_all     STRING = "ALL"
CONSTANT cs_doc_type_4rp     STRING = "4RP" 
CONSTANT cs_doc_type_xrp     STRING = "XRP" #230109-00005#1
CONSTANT cs_doc_type_code    STRING = "CODE"
CONSTANT cs_doc_type_cspec   STRING = "CSPEC"
CONSTANT cs_doc_type_gcode   STRING = "GCODE"
CONSTANT cs_doc_type_rspec   STRING = "RSPEC"
CONSTANT cs_doc_type_spec    STRING = "SPEC"
CONSTANT cs_doc_type_specgen STRING = "SPECGEN"
CONSTANT cs_doc_type_specrc3 STRING = "SPECRC3"
CONSTANT cs_doc_type_4rpupk  STRING = "4RPUPK" #180601-00012
CONSTANT cs_doc_type_xrpupk  STRING = "XRPUPK" #230109-00005#1

#170515-00013 add begin
CONSTANT cs_doc_type_c4gl    STRING = "C4GL"
CONSTANT cs_doc_type_c4fd    STRING = "C4FD"
#170515-00013 add end

CONSTANT cs_zip_name_tzc STRING = "tzc"
CONSTANT cs_zip_name_tzd STRING = "tzd"
CONSTANT cs_zip_name_tzg STRING = "tzg"
CONSTANT cs_zip_name_tzr STRING = "tzr"
CONSTANT cs_zip_name_tzs STRING = "tzs"
CONSTANT cs_zip_name_tzt STRING = "tzt"
CONSTANT cs_zip_name_tzp STRING = "tzp"

CONSTANT cs_zip_name_tzgt STRING = "tztt"

CONSTANT cs_upload_string STRING = "_up"

CONSTANT cs_client_temp_dir STRING = "c:\\t100temp"
CONSTANT cs_dos_separator   STRING = "\\"
CONSTANT cs_unix_separator  STRING = "/"

CONSTANT cs_open_4rp_dialog  STRING = "Open 4RP Directory" 
CONSTANT cs_open_xrp_dialog  STRING = "Open XRP Directory" #230109-00005#1

CONSTANT cs_message_designer       STRING = "   Designer   "
CONSTANT cs_message_generator      STRING = "  Generator   "    
CONSTANT cs_message_user_control   STRING = " User Control "
CONSTANT cs_message_separator      STRING = "================================================================================"
CONSTANT cs_message_half_separator STRING = "================================"
