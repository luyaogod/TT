#+ 程式版本......: T6 Version 1.00.00 Build-0001 at 15/01/05
#
#+ 程式代碼......: sadzp060_3
#+ 設計人員......: Hiko
# Prog. Version..: 'T6-12.01.21(00000)'     #
#
# Program name   : sadzp060_3.4gl
# Description    : 簽出備份相關功能
# Modify         : 2015/03/20 by Hiko : 新建程式
#                : 2015/11/17 by Hiko : 因為目前沒有版次歸一的議題,所以暫時取消備份機制,以加快下載速度.

import os

SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
&include "../4gl/sadzp060_type.inc"

DEFINE g_date DATETIME YEAR TO SECOND,
       gs_erpver STRING, #ERP大版版號
       gs_customer STRING, 
       g_env LIKE dzaa_t.dzaa009 #辨識目前所在的環境:s.產中環境,c.客製環境

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 初始化變數.
# Input parameter : none
# Return code     : BOOLEAN(成功:TRUE,失敗:FALSE)
# Date & Author   : 2015/01/05 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_3_init_var()
   DEFINE ls_trigger STRING,
          ls_err_msg STRING

   TRY
      LET ls_trigger = "call sadzp060_3_init_var start..."
      DISPLAY ls_trigger

      LET g_date = cl_get_current()
      LET gs_erpver = FGL_GETENV("ERPVER") CLIPPED
      LET gs_customer = FGL_GETENV("CUST") CLIPPED
      LET g_env = FGL_GETENV("DGENV") CLIPPED
      
      DISPLAY "call sadzp060_3_init_var finish!"

      RETURN TRUE
   CATCH
      LET ls_err_msg = "CALL sadzp060_3_init_var ERROR!"
      CALL sadzp060_3_err_catch(NULL, ls_err_msg)
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 擷取錯誤訊息.
# Input parameter : p_sql     執行的SQL
#                 : p_err_msg 錯誤訊息
# Return code     : void
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_3_err_catch(p_sql, p_err_msg)
   DEFINE p_sql     STRING,
          p_err_msg STRING

   DISPLAY "STATUS=",STATUS
   DISPLAY "SQLCA.SQLCODE=",SQLCA.SQLCODE
   DISPLAY "p_sql=",p_sql
   DISPLAY "p_err_msg=",p_err_msg

   INITIALIZE g_errparam TO NULL
   LET g_errparam.code = "!" 
   LET g_errparam.extend = p_err_msg
   LET g_errparam.popup = TRUE
   CALL cl_err()
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 簽出時要備份設計資料(包含離線檔(tzs/tzc/tzt))
# Input parameter : po_dzat 簽出資訊
# Return code     : BOOLEAN TRUE:成功/FALSE:失敗
# Date & Author   : 2015/01/05 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_3_backup_design_data(po_dzat)
   DEFINE po_dzat T_DZAT_PARA
   DEFINE ls_trigger STRING,
          ls_sql     STRING,
          li_cnt     SMALLINT,
          ls_err_msg STRING,
          lb_trans   BOOLEAN #是否觸發BEGIN WORK
   DEFINE l_program        LIKE dzat_t.dzat001, #程式代號(PK)
          l_revision       LIKE dzat_t.dzat002, #版次(PK)
          l_construct      LIKE dzat_t.dzat003, #建構類型(M/S/F/B/G/X..)
          l_check_out_type LIKE dzat_t.dzat004, #簽出類型(PK)(SD/PR)
          l_req_no         LIKE dzat_t.dzat005, #需求單號
          l_req_order      LIKE dzat_t.dzat006, #需求單項次
          l_date           LIKE dzat_t.dzat007, #簽出日期
          l_module         LIKE dzat_t.dzat008, #模組
          l_file           LIKE dzat_t.dzat009, #實體檔案(tzs/tzc)
          l_tzt            LIKE dzat_t.dzat010, #實體檔案(tzt)
          l_identity       LIKE dzat_t.dzat013, #客製(PK)
          l_old_rev        LIKE dzat_t.dzat002, #舊版次資料
          l_old_iden       LIKE dzat_t.dzat013, #舊客製資料
          ls_file_path     STRING,              #tzs/tzc的路徑(含副檔名)
          ls_tzt_path      STRING               #tzt的路徑(含副檔名):l_construct=G才需要
          
   RETURN TRUE #2015/11/17 by Hiko

   TRY
      IF NOT sadzp060_3_init_var() THEN
         RETURN FALSE
      END IF

      IF NOT sadzp060_3_check_param(po_dzat.*) THEN
         RETURN FALSE
      END IF

      LET ls_trigger = "call sadzp060_3_backup_design_data start..."
      DISPLAY ls_trigger

      LET l_program        = po_dzat.dzat001 CLIPPED
      LET l_revision       = po_dzat.dzat002
      LET l_construct      = po_dzat.dzat003 CLIPPED
      LET l_check_out_type = po_dzat.dzat004 CLIPPED
      LET l_req_no         = po_dzat.dzat005 CLIPPED
      LET l_req_order      = po_dzat.dzat006
      LET l_date           = po_dzat.dzat007
      LET l_module         = po_dzat.dzat008 CLIPPED
      LET l_identity       = po_dzat.dzat013 CLIPPED
      LET l_old_rev        = po_dzat.old_rev
      LET l_old_iden       = po_dzat.old_iden CLIPPED
      LET ls_file_path     = po_dzat.file_path CLIPPED
      LET ls_tzt_path      = po_dzat.tzt_path CLIPPED
      DISPLAY "l_program=",l_program
      DISPLAY "l_revision=",l_revision
      DISPLAY "l_construct=",l_construct
      DISPLAY "l_check_out_type=",l_check_out_type
      DISPLAY "l_req_no=",l_req_no
      DISPLAY "l_req_order=",l_req_order
      DISPLAY "l_date=",l_date
      DISPLAY "l_module=",l_module
      DISPLAY "l_identity=",l_identity
      DISPLAY "l_old_rev=",l_old_rev
      DISPLAY "l_old_iden=",l_old_iden
      DISPLAY "ls_file_path=",ls_file_path
      DISPLAY "ls_tzt_path=",ls_tzt_path

      LET lb_trans = FALSE
      BEGIN WORK
      LET lb_trans = TRUE

      #先找看看有沒有上一版的簽出資料:
      #有:將離線檔寫入上一版的資料內,這就類似上一版簽入時的備份.
      #最後再將本版次的簽出資料寫入備份檔內(不含離線檔),這是給下次簽出時準備.

      SELECT COUNT(*) INTO li_cnt FROM dzat_t
       WHERE dzat001=l_program AND dzat002=l_old_rev AND dzat004=l_check_out_type AND dzat013=l_old_iden

      IF li_cnt>0 THEN
         LOCATE l_file IN FILE
         LOCATE l_tzt IN FILE
         
         #有資料的話,應該就要有檔案路徑.
         IF cl_null(ls_file_path) THEN
            #DISPLAY "INFO : 程式 ",l_program," 沒有 ",l_check_out_type," 的離線檔案路徑."
            DISPLAY "INFO : The program ",l_program," did not find type ",l_check_out_type," file path."
         ELSE
            LET ls_trigger = "sadzp060_3_backup_design_data:read file ",ls_file_path
            DISPLAY ls_trigger
            IF os.path.exists(ls_file_path) then 
               CALL l_file.readFile(ls_file_path)
            ELSE
               #DISPLAY "INFO : 程式 ",l_program," 沒有 ",l_check_out_type," 的離線檔案路徑(",ls_file_path,")"
               DISPLAY "INFO : The program ",l_program," did not find type ",l_check_out_type," file path(",ls_file_path,")"
            END IF
         END IF
         
         IF l_construct="G" THEN #GR還得加上tzt路徑
            IF cl_null(ls_tzt_path) THEN
               #DISPLAY "INFO : 程式 ",l_program," 沒有 tzt 的離線檔案路徑."
               DISPLAY "INFO : The program ",l_program," did not find tzc file path."
            ELSE
               LET ls_trigger = "sadzp060_3_backup_design_data:read file ",ls_tzt_path
               DISPLAY ls_trigger
               IF os.path.exists(ls_tzt_path) then 
                  CALL l_tzt.readFile(ls_tzt_path)
               ELSE
                  #DISPLAY "INFO : 程式 ",l_program," 找不到 tzt 的離線檔案路徑(",ls_file_path,")"
                  DISPLAY "INFO : The program ",l_program," did not find tzc file path(",ls_file_path,")"
               END IF
            END IF
         END IF
         
         LET ls_trigger = "sadzp060_3_backup_design_data : update dzat_t data : old revision"
         DISPLAY ls_trigger
         UPDATE dzat_t SET dzat009=l_file,dzat010=l_tzt
          WHERE dzat001=l_program AND dzat002=l_old_rev AND dzat004=l_check_out_type AND dzat013=l_old_iden

         FREE l_tzt
         FREE l_file
      END IF      

      LET ls_trigger = "sadzp060_3_backup_design_data : insert dzat_t data : new revision"
      DISPLAY ls_trigger
      SELECT COUNT(*) INTO li_cnt FROM dzat_t
       WHERE dzat001=l_program AND dzat002=l_revision AND dzat004=l_check_out_type AND dzat013=l_identity

      IF li_cnt=0 THEN #正常來說,此FUNCTION只會在簽出時呼叫,因此一般不會出現重複問題,這裡只是防呆
         INSERT INTO dzat_t (dzat001,dzat002,dzat003,dzat004,dzat005,dzat006,
                             dzat007,dzat008,dzat009,dzat010,dzat011,dzat012,dzat013,
                             dzatcrtdt,dzatcrtdp,dzatowndp,dzatownid,dzatcrtid)
                     VALUES (l_program,l_revision,l_construct,l_check_out_type,l_req_no,l_req_order,
                             l_date,l_module,'','','','',l_identity,
                             g_date,g_dept,g_dept,g_user,g_user)
      END IF

      COMMIT WORK

      DISPLAY "call sadzp060_3_backup_design_data finish!"

      RETURN TRUE
   CATCH
      IF lb_trans THEN
         ROLLBACK WORK
      END IF

      LET ls_err_msg = "CALL sadzp060_3_backup_design_data ERROR!"
      CALL sadzp060_3_err_catch(ls_sql, ls_err_msg)
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 判斷參數是否正確
# Input parameter : po_dzat 設計資料備份檔資訊
# Return code     : BOOLEAN TRUE:成功/FALSE:失敗
# Date & Author   : 2015/01/05 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_3_check_param(po_dzat)
   DEFINE po_dzat T_DZAT_PARA
   DEFINE l_program        LIKE dzat_t.dzat001, #程式代號(PK)
          l_revision       LIKE dzat_t.dzat002, #版次(PK)
          l_check_out_type LIKE dzat_t.dzat004, #簽出類型(PK)(SD/PR)
          l_identity       LIKE dzat_t.dzat013, #客製標示(PK)
          ls_rep_str       STRING
          
   LET l_program        = po_dzat.dzat001 CLIPPED
   LET l_revision       = po_dzat.dzat002 
   LET l_check_out_type = po_dzat.dzat004 CLIPPED
   LET l_identity       = po_dzat.dzat013 CLIPPED
   
   IF cl_null(l_program) THEN
      LET ls_rep_str = "PROGRAM NAME"
      GOTO LBL_CHK_NULL_ERR
   END IF
   
   IF cl_null(l_revision) THEN
      LET ls_rep_str = "REVISION"
      GOTO LBL_CHK_NULL_ERR
   END IF
   
   IF cl_null(l_check_out_type) THEN
      LET ls_rep_str = "CHECK OUT TYPE"
      GOTO LBL_CHK_NULL_ERR
   END IF
   
   IF cl_null(l_identity) THEN
      LET ls_rep_str = "IDENTITY"
      GOTO LBL_CHK_NULL_ERR
   END IF
   
   RETURN TRUE
   
   LABEL LBL_CHK_NULL_ERR:
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = "adz-00263" #%1 資料為空 
      LET g_errparam.replace[1] = ls_rep_str
      LET g_errparam.popup = TRUE
      CALL cl_err()
      RETURN FALSE
END FUNCTION
