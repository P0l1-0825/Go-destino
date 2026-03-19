package domain

// Permission represents a granular access control permission.
type Permission string

// Reservation permissions.
const (
	PermResCreateWeb     Permission = "res.create.web"
	PermResCreateAPI     Permission = "res.create.api"
	PermResReadOwn       Permission = "res.read.own"
	PermResReadAirport   Permission = "res.read.airport"
	PermResReadCompany   Permission = "res.read.company"
	PermResReadAll       Permission = "res.read.all"
	PermResCancelOwn     Permission = "res.cancel.own"
	PermResCancelAny     Permission = "res.cancel.any"
	PermResAssignDriver  Permission = "res.assign.driver"
	PermResOverrideAI    Permission = "res.override.ai"
	PermResPriceEstimate Permission = "res.price.estimate"
)

// Payment permissions.
const (
	PermPayCharge        Permission = "pay.charge"
	PermPayCash          Permission = "pay.cash"
	PermPayQR            Permission = "pay.qr"
	PermPayRefundOwn     Permission = "pay.refund.own"
	PermPayRefundAny     Permission = "pay.refund.any"
	PermPayMethodsManage Permission = "pay.methods.manage"
	PermPayVoucherCreate Permission = "pay.voucher.create"
	PermPayVoucherRedeem Permission = "pay.voucher.redeem"
	PermPayReportOwn     Permission = "pay.report.own"
	PermPayReportCompany Permission = "pay.report.company"
	PermPayReportGlobal  Permission = "pay.report.global"
	PermPayLiquidation   Permission = "pay.liquidation"
	PermPayInvoice       Permission = "pay.invoice"
	PermPayGatewayConfig Permission = "pay.gateway.config"
	PermPayExportFiscal  Permission = "pay.export.fiscal"
)

// Fleet permissions.
const (
	PermFleetLocationOwn   Permission = "fleet.location.own"
	PermFleetLocationView  Permission = "fleet.location.view"
	PermFleetStatusOwn     Permission = "fleet.status.own"
	PermFleetDriverOnboard Permission = "fleet.driver.onboard"
	PermFleetDriverRead    Permission = "fleet.driver.read"
	PermFleetDriverManage  Permission = "fleet.driver.manage"
	PermFleetDriverVerify  Permission = "fleet.driver.verify"
	PermFleetDriverRate    Permission = "fleet.driver.rate"
	PermFleetVehicleOwn    Permission = "fleet.vehicle.own"
	PermFleetVehicleAll    Permission = "fleet.vehicle.all"
	PermFleetDispatchMap   Permission = "fleet.dispatch.map"
	PermFleetHeatmap       Permission = "fleet.heatmap"
)

// Analytics permissions.
const (
	PermAnalyticsKPIBasic   Permission = "analytics.kpi.basic"
	PermAnalyticsKPIAirport Permission = "analytics.kpi.airport"
	PermAnalyticsKPICompany Permission = "analytics.kpi.company"
	PermAnalyticsKPIGlobal  Permission = "analytics.kpi.global"
	PermAnalyticsReports    Permission = "analytics.reports"
	PermAnalyticsExport     Permission = "analytics.export"
	PermAnalyticsCohort     Permission = "analytics.cohort"
	PermAnalyticsSLO        Permission = "analytics.slo"
)

// AI permissions.
const (
	PermAIChat           Permission = "ai.chat"
	PermAIInsightsView   Permission = "ai.insights.view"
	PermAIPricingView    Permission = "ai.pricing.view"
	PermAIPricingConfig  Permission = "ai.pricing.config"
	PermAIDemandForecast Permission = "ai.demand.forecast"
	PermAIFraudAlerts    Permission = "ai.fraud.alerts"
	PermAIModelsRetrain  Permission = "ai.models.retrain"
)

// System permissions.
const (
	PermSysUsersRead     Permission = "sys.users.read"
	PermSysUsersManage   Permission = "sys.users.manage"
	PermSysRolesAssign   Permission = "sys.roles.assign"
	PermSysRolesCreate   Permission = "sys.roles.create"
	PermSysAirportsRead  Permission = "sys.airports.read"
	PermSysAirportsManage Permission = "sys.airports.manage"
	PermSysKioskView     Permission = "sys.kiosk.view"
	PermSysKioskManage   Permission = "sys.kiosk.manage"
	PermSysSettingsView  Permission = "sys.settings.view"
	PermSysSettingsEdit  Permission = "sys.settings.edit"
	PermSysAuditLog      Permission = "sys.audit.log"
	PermSysAPIKeys       Permission = "sys.api.keys"
	PermSysWebhooks      Permission = "sys.webhooks"
)

// Kiosk/POS operations permissions.
const (
	PermKioskBookCreate    Permission = "kiosk.book.create"
	PermKioskPrintTicket   Permission = "kiosk.print.ticket"
	PermKioskOfflineSync   Permission = "kiosk.offline.sync"
	PermKioskShiftOpen     Permission = "kiosk.shift.open"
	PermKioskShiftClose    Permission = "kiosk.shift.close"
	PermKioskCommissionView Permission = "kiosk.commission.view"
)

// Concesion permissions.
const (
	PermConcesionRead       Permission = "concesion.read"
	PermConcesionReadOwn    Permission = "concesion.read.own"    // read own concesion only
	PermConcesionCreate     Permission = "concesion.create"
	PermConcesionUpdate     Permission = "concesion.update"
	PermConcesionUpdateOwn  Permission = "concesion.update.own"  // update own concesion
	PermConcesionDelete     Permission = "concesion.delete"
	PermConcesionStaffRead  Permission = "concesion.staff.read"  // list staff in concesion
	PermConcesionStaffManage Permission = "concesion.staff.manage" // assign/remove staff
	PermConcesionVehicleRead Permission = "concesion.vehicle.read"
	PermConcesionVehicleManage Permission = "concesion.vehicle.manage"
)

// QR validation permissions.
const (
	PermQRValidate Permission = "qr.validate"
)

// AllPermissions returns every permission defined in the system.
func AllPermissions() []Permission {
	return []Permission{
		PermResCreateWeb, PermResCreateAPI, PermResReadOwn, PermResReadAirport,
		PermResReadCompany, PermResReadAll, PermResCancelOwn, PermResCancelAny,
		PermResAssignDriver, PermResOverrideAI, PermResPriceEstimate,
		PermPayCharge, PermPayCash, PermPayQR, PermPayRefundOwn, PermPayRefundAny,
		PermPayMethodsManage, PermPayVoucherCreate, PermPayVoucherRedeem,
		PermPayReportOwn, PermPayReportCompany, PermPayReportGlobal,
		PermPayLiquidation, PermPayInvoice, PermPayGatewayConfig, PermPayExportFiscal,
		PermFleetLocationOwn, PermFleetLocationView, PermFleetStatusOwn,
		PermFleetDriverOnboard, PermFleetDriverRead, PermFleetDriverManage,
		PermFleetDriverVerify, PermFleetDriverRate, PermFleetVehicleOwn,
		PermFleetVehicleAll, PermFleetDispatchMap, PermFleetHeatmap,
		PermAnalyticsKPIBasic, PermAnalyticsKPIAirport, PermAnalyticsKPICompany,
		PermAnalyticsKPIGlobal, PermAnalyticsReports, PermAnalyticsExport,
		PermAnalyticsCohort, PermAnalyticsSLO,
		PermAIChat, PermAIInsightsView, PermAIPricingView, PermAIPricingConfig,
		PermAIDemandForecast, PermAIFraudAlerts, PermAIModelsRetrain,
		PermSysUsersRead, PermSysUsersManage, PermSysRolesAssign, PermSysRolesCreate,
		PermSysAirportsRead, PermSysAirportsManage, PermSysKioskView, PermSysKioskManage,
		PermSysSettingsView, PermSysSettingsEdit, PermSysAuditLog, PermSysAPIKeys, PermSysWebhooks,
		PermKioskBookCreate, PermKioskPrintTicket, PermKioskOfflineSync,
		PermKioskShiftOpen, PermKioskShiftClose, PermKioskCommissionView,
		PermConcesionRead, PermConcesionReadOwn, PermConcesionCreate,
		PermConcesionUpdate, PermConcesionUpdateOwn, PermConcesionDelete,
		PermConcesionStaffRead, PermConcesionStaffManage,
		PermConcesionVehicleRead, PermConcesionVehicleManage,
		PermQRValidate,
	}
}

// RolePermissions maps each role to its allowed permissions.
var RolePermissions = map[UserRole][]Permission{
	RoleSuperAdmin: AllPermissions(),
	RoleAdmin: {
		PermResCreateWeb, PermResReadAirport, PermResReadAll, PermResCancelAny,
		PermResAssignDriver, PermResOverrideAI, PermResPriceEstimate,
		PermPayRefundAny, PermPayReportCompany, PermPayReportGlobal,
		PermPayGatewayConfig, PermPayExportFiscal,
		PermFleetLocationView, PermFleetDriverOnboard, PermFleetDriverRead,
		PermFleetDriverManage, PermFleetDriverVerify, PermFleetVehicleAll,
		PermFleetDispatchMap, PermFleetHeatmap,
		PermAnalyticsKPIAirport, PermAnalyticsKPIGlobal, PermAnalyticsReports,
		PermAnalyticsExport, PermAnalyticsCohort, PermAnalyticsSLO,
		PermAIInsightsView, PermAIPricingView, PermAIDemandForecast, PermAIFraudAlerts,
		PermSysUsersRead, PermSysUsersManage, PermSysRolesAssign,
		PermSysAirportsRead, PermSysAirportsManage, PermSysKioskView,
		PermSysKioskManage, PermSysSettingsView, PermSysAuditLog, PermSysAPIKeys,
		PermKioskBookCreate, PermKioskOfflineSync,
		// Concesion: admin manages all concesiones
		PermConcesionRead, PermConcesionCreate, PermConcesionUpdate, PermConcesionDelete,
		PermConcesionStaffRead, PermConcesionStaffManage,
		PermConcesionVehicleRead, PermConcesionVehicleManage,
	},
	RoleClienteConcesion: {
		PermResReadCompany, PermResPriceEstimate,
		PermPayReportCompany, PermPayLiquidation, PermPayInvoice, PermPayExportFiscal,
		PermFleetDriverOnboard, PermFleetDriverRead, PermFleetDriverManage,
		PermFleetDriverVerify, PermFleetVehicleOwn, PermFleetVehicleAll,
		PermAnalyticsKPICompany, PermAnalyticsReports, PermAnalyticsExport,
		PermSysKioskView, PermSysSettingsView,
		// Concesion: administrativo manages own concesion
		PermConcesionReadOwn, PermConcesionUpdateOwn,
		PermConcesionStaffRead, PermConcesionStaffManage,
		PermConcesionVehicleRead, PermConcesionVehicleManage,
	},
	RoleTesoreriaCliente: {
		PermResReadCompany,
		PermPayReportCompany, PermPayLiquidation, PermPayInvoice,
		PermPayExportFiscal, PermPayRefundOwn,
		PermAnalyticsKPICompany, PermAnalyticsReports, PermAnalyticsExport,
		PermConcesionReadOwn, // read own concesion for financial context
	},
	RoleMesaControl: {
		PermResReadAirport, PermResCancelAny, PermResAssignDriver, PermResOverrideAI,
		PermFleetLocationView, PermFleetDriverRead, PermFleetDispatchMap, PermFleetHeatmap,
		PermAnalyticsKPIAirport, PermAnalyticsKPIBasic,
		PermAIInsightsView, PermAIDemandForecast,
		PermSysKioskView,
		PermQRValidate,
	},
	RoleOperador: {
		PermResReadAirport,
		PermFleetDriverRead, PermFleetDriverVerify, PermFleetLocationView, PermFleetDispatchMap,
		PermAnalyticsKPIBasic,
		PermSysKioskView,
		PermConcesionReadOwn, PermConcesionStaffRead, PermConcesionVehicleRead, // operativo reads concesion
	},
	RoleTaxista: {
		PermResReadOwn, PermResCancelOwn,
		PermFleetLocationOwn, PermFleetStatusOwn, PermFleetVehicleOwn,
		PermAnalyticsKPIBasic,
		PermAIChat,
		PermQRValidate,
		PermConcesionReadOwn, // taxista sees own concesion info
	},
	RoleVendedor: {
		PermResCreateWeb, PermResReadAirport, PermResCancelOwn, PermResPriceEstimate,
		PermPayCharge, PermPayCash, PermPayQR, PermPayVoucherCreate,
		PermPayVoucherRedeem, PermPayReportOwn,
		PermAnalyticsKPIBasic,
		PermKioskBookCreate, PermKioskPrintTicket, PermKioskShiftOpen,
		PermKioskShiftClose, PermKioskCommissionView,
		PermQRValidate,
	},
	RoleBroker: {
		PermResCreateAPI, PermResReadOwn, PermResCancelOwn, PermResPriceEstimate,
		PermPayMethodsManage, PermPayReportOwn,
		PermAnalyticsKPIBasic,
		PermAIChat,
		PermSysAPIKeys, PermSysWebhooks,
	},
	RoleUsuario: {
		PermResCreateWeb, PermResReadOwn, PermResCancelOwn, PermResPriceEstimate,
		PermPayCharge, PermPayMethodsManage, PermPayRefundOwn,
		PermFleetDriverRate,
		PermAIChat,
	},
}

// HasPermission checks if a role has a specific permission.
func HasPermission(role UserRole, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
