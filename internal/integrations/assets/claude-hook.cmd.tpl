@echo off
rem agentusage-integration-version: __AGENTUSAGE_INTEGRATION_VERSION__
if /I "%AGENTUSAGE_TELEMETRY_ENABLED%"=="false" exit /b 0
if /I "%AGENTUSAGE_TELEMETRY_ENABLED%"=="0" exit /b 0
"__AGENTUSAGE_BIN_DEFAULT__" telemetry hook claude_code 1>nul 2>nul
