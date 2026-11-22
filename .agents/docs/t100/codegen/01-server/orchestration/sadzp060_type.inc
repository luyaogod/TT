PRIVATE TYPE T_DZAT_PARA RECORD
               DZAT001   LIKE DZAT_T.DZAT001, #程式代號(PK)
               DZAT002   LIKE DZAT_T.DZAT002, #版次(PK)
               DZAT003   LIKE DZAT_T.DZAT003, #建構類型
               DZAT004   LIKE DZAT_T.DZAT004, #簽出類型:SD/PR(PK)
               DZAT005   LIKE DZAT_T.DZAT005, #需求單號
               DZAT006   LIKE DZAT_T.DZAT006, #需求單項次
               DZAT007   LIKE DZAT_T.DZAT007, #簽出日期
               DZAT008   LIKE DZAT_T.DZAT008, #模組
               DZAT013   LIKE DZAT_T.DZAT013, #客製(PK)
               OLD_REV   LIKE DZAT_T.DZAT002, #舊版次
               OLD_IDEN  LIKE DZAT_T.DZAT013, #舊客製
               FILE_PATH VARCHAR(1024),       #TZS/TZC的路徑
               TZT_PATH  VARCHAR(1024)        #TZT的路徑,DZAT003=G才需要
             END RECORD  

