param([int]$TargetPid, [string]$AutoIdSuffix)
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
$root = [System.Windows.Automation.AutomationElement]::RootElement
$cond = New-Object System.Windows.Automation.PropertyCondition([System.Windows.Automation.AutomationElement]::ProcessIdProperty, $TargetPid)
$els = $root.FindAll([System.Windows.Automation.TreeScope]::Descendants, $cond)
foreach ($e in $els) {
  $id = $e.Current.AutomationId
  if ($id -eq $null -or -not $id.EndsWith($AutoIdSuffix)) { continue }
  $vp = $null
  if ($e.TryGetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern, [ref]$vp)) {
    Write-Output ($id + " = " + $vp.Current.Value); exit 0
  }
  $tp = $null
  if ($e.TryGetCurrentPattern([System.Windows.Automation.TogglePattern]::Pattern, [ref]$tp)) {
    Write-Output ($id + " = " + $tp.Current.ToggleState); exit 0
  }
  Write-Output ($id + " = (no value pattern; type=" + $e.Current.ControlType.ProgrammaticName + ")"); exit 0
}
Write-Output "NOT_FOUND $AutoIdSuffix"; exit 1
