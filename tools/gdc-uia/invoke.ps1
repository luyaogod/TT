param([int]$TargetPid, [string]$AutoIdSuffix)
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
$p = $hit.GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern)
$p.Invoke()
Write-Output ("INVOKED " + $hit.Current.AutomationId)
