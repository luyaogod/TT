param([int]$TargetPid, [string]$AutoIdSuffix, [string]$Value, [string]$Type = "")
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
$root = [System.Windows.Automation.AutomationElement]::RootElement
$cond = New-Object System.Windows.Automation.PropertyCondition([System.Windows.Automation.AutomationElement]::ProcessIdProperty, $TargetPid)
$els = $root.FindAll([System.Windows.Automation.TreeScope]::Descendants, $cond)
$hit = $null
foreach ($e in $els) {
  $id = $e.Current.AutomationId
  if (($id -ne $null -and $id.EndsWith($AutoIdSuffix)) -or ($e.Current.Name -eq $AutoIdSuffix)) {
    if ($Type -ne "" -and $e.Current.ControlType.ProgrammaticName -notlike "*$Type") { continue }
    $hit = $e; break
  }
}
if ($hit -eq $null) { Write-Output "NOT_FOUND $AutoIdSuffix"; exit 1 }
$tp0 = $null
if ($hit.TryGetCurrentPattern([System.Windows.Automation.TogglePattern]::Pattern, [ref]$tp0)) {
  $tp0.Toggle()
  Write-Output ("TOGGLED " + $hit.Current.Name + " (now " + $tp0.Current.ToggleState + ")")
  exit 0
}
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
