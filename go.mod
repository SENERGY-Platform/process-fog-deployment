module github.com/SENERGY-Platform/process-fog-deployment

go 1.16

require (
	github.com/SENERGY-Platform/process-deployment v0.0.0-20210312065838-c5ece7c3f107
	github.com/SmartEnergyPlatform/jwt-http-router v0.0.0-20190722084820-0e1fe0dc7a07
	github.com/julienschmidt/httprouter v1.3.0
)

//replace github.com/SENERGY-Platform/process-deployment => ../process-deployment