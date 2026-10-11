param([int]$TargetPid, [string]$AutoIdSuffix, [string]$Value)
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
$root = [System.Windows.Automation.AutomationElement]::RootElement
$cond = New-Object System.Windows.Automation.PropertyCondition([System.Windows.Automation.AutomationElement]::ProcessIdProperty, $TargetPid)
$els = $root.FindAll([System.Windows.Automation.TreeScope]::Descendants, $cond)
$hit = $null
foreach ($e in $els) {
  $id = $e.Current.AutomationId
  if ($id -ne $null -and $id.EndsWith($AutoIdSuffix)) { $hit = $e; break }
}
if ($hit -eq $null) { Write-Output "NOT_FOUND $AutoIdSuffix"; exit 1 }
$vp = $null
if ($hit.TryGetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern, [ref]$vp)) {
  $vp.SetValue($Value)
  Write-Output ("SETVALUE " + $hit.Current.AutomationId + " = " + $Value)
  exit 0
}
$tp = $null
if ($hit.TryGetCurrentPattern([System.Windows.Automation.TogglePattern]::Pattern, [ref]$tp)) {
  $tp.Toggle()
  Write-Output ("TOGGLED " + $hit.Current.AutomationId + " (now " + $tp.Current.ToggleState + ")")
  exit 0
}
Write-Output ("NO_PATTERN " + $hit.Current.AutomationId + " type=" + $hit.Current.ControlType.ProgrammaticName)
exit 2
