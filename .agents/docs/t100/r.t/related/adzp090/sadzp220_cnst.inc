#Argument Types
CONSTANT cs_args_working_type         STRING = "-WT"
CONSTANT cs_args_source_full_name     STRING = "-SFN"
CONSTANT cs_args_export_file_location STRING = "-EFL"
CONSTANT cs_args_execute_guid         STRING = "-GUID"

#Working type
CONSTANT cs_state_holding    = "H" -- Holding 
CONSTANT cs_state_waiting    = "W" -- Waiting
CONSTANT cs_state_assigned   = "A" -- Assigned
CONSTANT cs_state_processing = "P" -- Processing
#Finished type
CONSTANT cs_state_finished   = "F" -- Finished
CONSTANT cs_state_error      = "E" -- Error
CONSTANT cs_state_manually   = "M" -- Manually
CONSTANT cs_state_cancel     = "X" -- Cancel
CONSTANT cs_state_invalid    = "I" -- Invalid 
