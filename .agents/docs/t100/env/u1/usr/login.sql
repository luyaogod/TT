--
-- Copyright (c) 1988, 2011, Oracle and/or its affiliates. 
-- All rights reserved. 
--
-- NAME
--   glogin.sql
--
-- DESCRIPTION
--   SQL*Plus global login "site profile" file
--
--   Add any SQL*Plus commands here that are to be executed when a
--   user starts SQL*Plus, or uses the SQL*Plus CONNECT command.
--
-- USAGE
--   This script is automatically run
--
-- DSC: show environment --
define gname=idle  
column global_name new_value gname  
select sys_context('userenv', 'host')||'/'||lower(user) || '@' ||  
       substr( global_name, 1, decode( dot,  
                                      0, length(global_name),  
                                          dot-1) ) global_name  
  from (select global_name, instr(global_name,'.') dot  
          from global_name );  
--set sqlprompt '&gname> '  

-- DSC: alter session --
DECLARE
  V_A VARCHAR2(10);
BEGIN
  select decode((select sys_context('USERENV','SESSION_USER') current_user from dual), 'DU', 'DS', 'DUDEMO', 'DSDEMO' ,'DUDATA', 'DSDATA', (select sys_context('USERENV','SESSION_USER') from dual)) into V_A from dual;
  execute immediate 'alter session set current_schema='||V_A;
END; 
/
