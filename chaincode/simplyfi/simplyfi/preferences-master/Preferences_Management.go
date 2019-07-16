/*
Copyright Tanla Solutions Ltd. 2019 All Rights Reserved.
This Chaincode is written for storing,retrieving,updating,
deleting(churnout) the preferences that are stored in DLT
and portOut the MSISDN on Successful Certificate verification.
*/
package main

import (
        "encoding/json"                                        //reading and writing JSON
        id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid" // import for Client Identity
        "github.com/hyperledger/fabric/core/chaincode/shim"    // import for Chaincode Interface
        pb "github.com/hyperledger/fabric/protos/peer"         // import for peer response
        "strconv"                                              //import for msisdn validation
        "strings"
)

//Logger for Logging
var _preferencesLogger = shim.NewLogger("Preferences-Manager")

//Event Names
const _AddEvent      = "ADD_PREFERENCES"
const _UpdateEvent   = "UPDATE_PREFERENCES"
const _PortOutEvent  = "PORT_OUT"
const _DeleteEvent   = "DELETE_PREFERENCES"
const _SnapBackEvent = "SNAP_BACK_CHURN"


//=========================================================================================================
// Preference structure, with 17 properties.  Structure tags are used by encoding/json library
//=========================================================================================================
type Preference struct {
        ObjType            string `json:"obj"`
        Phone              string `json:"msisdn"`
        ServiceProvider    string `json:"svcprv"`
        RequestNumber      string `json:"reqno"`
        RegistrationMode   string `json:"rmode"`
        Category           string `json:"ctgr"`
        CommunicationMode  string `json:"cmode"`
        DayType            string `json:"day"`
        DayTimeBand        string `json:"time"`
        Lrn                string `json:"lrn"`
	CreateTs	   string `json:"cts"`
        UpdateTs           string `json:"uts"`
	Creator	           string `json:"crtr"`
        UpdatedBy          string `json:"uby"`
	CRMReferenceNumber string `json:"crmno,omitempty"`
	Status		   string `json:"sts"`
	ServiceAreaCode    string `json:"srvac"`
	PhoneType          string `json:"ptype,omitempty"`
}



//Smart Contract structure
type PreferencesManager struct {
}


var serviceProviders = map[string]bool{
        "AI": true,//Airtel
        "VO": true,//Vodafone
        "ID": true,//IDEA
        "BL": true,//BSNL
        "ML": true,//MTNL
        "QL": true,//QTL
        "TA": true,//TATA
        "JI": true,//JIO
        "VI": true,//Vodafone Idea DLT
}

var dltDomainNames = map [string]string{
	"AI":"airtel.com",//Airtel
	"VO":"vil.com",//Vodafone
	"ID":"vil.com",//IDEA
	"BL":"bsnl.com",//BSNL
        "ML":"mtnl.com",//MTNL
        "QL":"qtl.infotelconnect.com",//QTL
        "TA":"tata.com",//TATA
        "JI":"jio.com",//JIO
        "VI":"vil.com",//Vodafone Idea DLT
}

var registrationModes= map[string]bool{
	"0":true, //Migration
	"1":true, //WEB 
	"2":true, //SMS
	"3":true, //IVR
	"4":true, //USSD
	"5":true, //APP
	"6":true, //CS
}

var statusCheck = map[string]bool{
        "A":true,//Active
        "T":true,//Terminated
	"D":true,//Deactivated
}


var serviceAreaCodes = map[string]bool{
	"1":true, //Andhra Pradesh
	"2":true, //Assam
	"3":true, //Bihar
	"4":true, //Chennai
	"5":true, //Delhi
	"6":true, //Gujarat
	"7":true, //Haryana
	"8":true, //Himachal Prade
	"9":true, //Jammu & Kashmi
	"10":true,//Karnataka
	"11":true,//Kerala
	"12":true,//Kolkata
	"13":true,//MadhyaPradesh
	"14":true,//Maharastra
	"15":true,//Mumbai
	"16":true,//North East
	"17":true,//Orissa
	"18":true,//Punjab
	"19":true,//Rajasthan
	"20":true,//Tamilnadu
	"21":true,//UP East
	"22":true,//UP West
	"23":true,//West Bengal
}


var phoneTypes = map[string]bool{
	"1":true,//Landline
	"2":true,//Mobile
	"3":true,//Other
}



var jsonResp string
var errorKey string
var errorData string

func validPreferencesEntry(input string, enumMap map[string]bool) bool {
        if _, isEntryExists := enumMap[input]; !isEntryExists {
                return false
        }
        return true
}

func isValidPreferences(pref Preference) (bool ,string){
        if len(pref.Phone)<10{
                errorKey=string(pref.Phone)
                errorData="Invalid Msisdn Length "
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false , string(jsonResp)
        }
        if _,err:=strconv.Atoi(pref.Phone); err!=nil{
                errorKey=string(pref.Phone)
                errorData="Msisdn is not Numeric"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
        if len(pref.ServiceProvider)==0{
                errorKey=string(pref.ServiceProvider)
                errorData="ServiceProvider is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
        if !validPreferencesEntry(pref.ServiceProvider,serviceProviders){
                errorKey=string(pref.ServiceProvider)
                errorData="Invalid ServiceProvider"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
	if len(pref.ServiceAreaCode)==0{
                errorKey=string(pref.ServiceAreaCode)
                errorData="ServiceAreaCode  is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
        if !validPreferencesEntry(pref.ServiceAreaCode,serviceAreaCodes){
                errorKey=string(pref.ServiceAreaCode)
                errorData="Invalid ServiceAreaCode"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
        if len(pref.RequestNumber)==0{
                errorKey=string(pref.RequestNumber)
                errorData="RequestNumber is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
	if len(pref.RegistrationMode)==0{
                errorKey=string(pref.RegistrationMode)
                errorData="RegistrationMode is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
	}
	if !validPreferencesEntry(pref.RegistrationMode,registrationModes){
                errorKey=string(pref.RegistrationMode)
                errorData="Invalid RegistrationMode "
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
	}
        if len(pref.Lrn)!=4{
                errorKey=string(pref.Lrn)
                errorData="Invalid Lrn Length"
		jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
        if _,err:=strconv.Atoi(pref.Lrn);err!=nil{
                errorKey=string(pref.Lrn)
                errorData="Lrn is not Numeric"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
	if len(pref.CreateTs)==0{
                errorKey=string(pref.CreateTs)
                errorData="CreateTs is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
	}
        if len(pref.UpdateTs)==0{
                errorKey=string(pref.UpdateTs)
                errorData="UpdateTs is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
	if len(pref.Creator)==0{
                errorKey=string(pref.Creator)
                errorData="Creator is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
	}
	if len(pref.UpdatedBy)==0{
                errorKey=string(pref.UpdatedBy)
                errorData="UpdatedBy is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
	}
	if len(pref.Status)==0{
                errorKey=string(pref.Status)
                errorData="Status is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
	if len(pref.Status)>0{
		if pref.Status!="T"{
			if !validPreferencesEntry(pref.Status,statusCheck){
		                errorKey=string(pref.Status)
				errorData="Invalid Status"
				jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
				_preferencesLogger.Error(string(jsonResp))
				return  false ,string(jsonResp)
			}
		}else{
			errorKey=string(pref.Status)
			errorData="Invalid Status"
			jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
			_preferencesLogger.Error(string(jsonResp))
			return false, string(jsonResp)
		}
        }
	if len(pref.PhoneType)>0{
		 if !validPreferencesEntry(pref.PhoneType,phoneTypes){
			errorKey=string(pref.PhoneType)
			errorData="Invalid PhoneType "
			jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
			_preferencesLogger.Error(string(jsonResp))
			return false, string(jsonResp)
	        }
	}
        return true,""
}




//======================================================================================================
//isValidParameters Checks if the Preferences portout fields are valid or not Or snapChurn fields are valid or not
//====================================================================================================== 
func isValidParameters(pref Preference) (bool ,string){
         if len(pref.ServiceProvider)==0{
                errorKey=string(pref.ServiceProvider)
                errorData="ServiceProvider is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false,string(jsonResp)
        }
        if !validPreferencesEntry(pref.ServiceProvider,serviceProviders){
                errorKey=string(pref.ServiceProvider)
                errorData="Invalid ServiceProvider"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false,string(jsonResp)
        }
	if len(pref.ServiceAreaCode)==0{
                errorKey=string(pref.ServiceAreaCode)
                errorData="ServiceAreaCode  is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }
        if !validPreferencesEntry(pref.ServiceAreaCode,serviceAreaCodes){
                errorKey=string(pref.ServiceAreaCode)
                errorData="Invalid ServiceAreaCode"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }

        if len(pref.Lrn)!=4{
                errorKey=string(pref.Lrn)
                errorData="Invalid Lrn Length "
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false,string(jsonResp)
        }
        if _,err:=strconv.Atoi(pref.Lrn);err!=nil{
                errorKey=string(pref.Lrn)
                errorData="Lrn is not Numeric"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false,string(jsonResp)
        }
        if len(pref.UpdateTs)==0{
                errorKey=string(pref.UpdateTs)
                errorData="UpdateTs is Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false,string(jsonResp)
        }
	if len(pref.UpdatedBy)==0{
                errorKey=string(pref.UpdatedBy)
                errorData="UpdatedBy is  Mandatory"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false,string(jsonResp)

	}
	if !validPreferencesEntry(pref.Status,statusCheck){
                errorKey=string(pref.Status)
                errorData="Invalid Status"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error(string(jsonResp))
                return false, string(jsonResp)
        }

        return true,""

}



//Returns the complete identity in the format
//Certitificate issuer orgs's domain name
//Returns string Unkown if not able parse the invoker certificate
func (pm *PreferencesManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool,string){
 //Following id comes in the format X509::<Subject>::<Issuer>>
        enCert, err := id.GetX509Certificate(stub)
        if err != nil {
                _preferencesLogger.Errorf("Getting Certificate Details Failed:"+string(err.Error()))
                return false, "Unknown"
        }
        issuersOrgs := enCert.Issuer.Organization
        if len(issuersOrgs)==0{
                return false,"Unknown"
        }
        domainName:=issuersOrgs[0]
        return true, string(domainName)
}



//=========================================================================================================
// The Init method is called when the Smart Contract "Preferences" is instantiated by the blockchain network
//=========================================================================================================
func (pm *PreferencesManager) Init(stub shim.ChaincodeStubInterface) pb.Response {
        _preferencesLogger.Info("###### Preferences-Chaincode is Initialized #######")
        return shim.Success(nil)
}


func (pm *PreferencesManager) Invoke(stub shim.ChaincodeStubInterface) pb.Response{
        action,args:=stub.GetFunctionAndParameters()
        _preferencesLogger.Infof("Preferences ChainCode is Invoked with Action Name is : " + string(action))
        switch action{
                case "sp"://add preferences into DL
                        return pm.setPreferences(stub,args)
		case "po"://Ownership transfer from donor to acceptor
                        return pm.portOut(stub,args)
		case "dp"://churnout  Preferences From DL
                        return pm.deletePreferences(stub,args)
		case "sbc"://Ownereship transfer from acceptor to donot
			return pm.snapBackChurn(stub,args)
		case "abp"://add/update Bulk preferences into DL
                        return pm.batchPreferences(stub,args)
		case "bpo"://Bulkwise owner ship transfer from donor to acceptor
                        return pm.batchPortOut(stub,args)
		case "dbp"://Bulk churnouts  from DL
                        return pm.batchDeletePreferences(stub,args)
		case "bsbc"://Bulk SnapBackChurn from DL
			return pm.batchSnapBackChurn(stub,args)
		case "pd"://getPreferences Details from dlt based on msisdn
			return pm.getPreferencesByMsisdn(stub,args)
                case "qp"://Rich Query to retrieve the Preferences from DL
                        return pm.queryPreferences(stub,args)
                case "hp"://get Preferences History from DL
                        return pm.getHistoryPreferences(stub,args)
		case "qpp"://query Preferences with pagination
			return pm.queryPreferencesWithPagination(stub,args)
                default:
                        _preferencesLogger.Errorf("Unknown Function Invoked, Available Functions : sp,po,dp,sbc,abp,bpo,dbp,bsbc,pd,qp,hp,qpp")
			jsonResp="{\"Data\":"+action+",\"ErrorDetails\":\"Available Functions:sp,po,dp,sbc,abp,bpo,dbp,bsbc,p,qp,hp,qpp\"}"
                        return shim.Error(jsonResp)
        }
}




//=================================================
//setPreferences - Setting new preference into dlt
// ================================================
func  (pm *PreferencesManager) setPreferences(stub shim.ChaincodeStubInterface,args []string)pb.Response{
        if len(args)!=1{
                _preferencesLogger.Errorf("setPreferences:Invalid Number of arguments provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
	var prefObj Preference
        err:=json.Unmarshal([]byte(args[0]),&prefObj)
        if err!=nil{
                errorKey=args[0]
                errorData="Invalid json provided as input"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("setPreferences:"+string(jsonResp))
                return shim.Error(jsonResp)
        }
        _,creator:=pm.getInvokerIdentity(stub)
        preferencesExist,err:=stub.GetState(prefObj.Phone)
        if err!=nil{
                errorKey=prefObj.Phone
		replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                errorData="GetState is Failed :"+replaceErr
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("setPreferences:"+string(jsonResp))
                return shim.Error(jsonResp)
        }
        if preferencesExist==nil{
                prefObj.ObjType="Preferences"
		prefObj.Creator=creator
                prefObj.UpdatedBy=creator
                preferencesJson,err:=json.Marshal(prefObj)
                if err!=nil{
                        _preferencesLogger.Errorf("setPreferences : Marshalling Error : " + string(err.Error()))
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Marshalling Error :"+replaceErr
			jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
                }
		if isValid,errMsg:=isValidPreferences(prefObj);!isValid{
                        return shim.Error(errMsg)
                }
                err=stub.PutState(prefObj.Phone,preferencesJson)
                if err!=nil{
                        _preferencesLogger.Errorf("setPreferences:PutState is Failed :"+string(err.Error()))
			jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\"Unable to set the Preferences\"}"
                        return shim.Error(jsonResp)
                }
		_preferencesLogger.Infof("setPreferences:Preferences added succesfull for Msisdn is :"+string(prefObj.Phone))
                err=stub.SetEvent(_AddEvent,[]byte(string(preferencesJson)))
                if err!=nil{
                        _preferencesLogger.Errorf("setPreferences:Event Not Generated for Event is:"+string(_AddEvent)+"Error is :"+string(err.Error()) )
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Event is Not Generated "+replaceErr
			jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
                }
		_preferencesLogger.Infof("setPreferences:Event Payload is :"+string(preferencesJson))
        }else{
		var updatedBy string
                preference:=Preference{}
                err:=json.Unmarshal(preferencesExist,&preference)
                if err!=nil{
                        _preferencesLogger.Errorf("setPreferences:Existing PreferencesData unmarshalling Error :"+string(err.Error()))
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                        errorData="Unmarshalling Error :"+replaceErr
                        jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)

                }
                updatedBy=preference.UpdatedBy
		if strings.Compare(updatedBy,creator)==0{
			prefObj.ObjType="Preferences"
			prefObj.CreateTs=preference.CreateTs
			prefObj.Creator=preference.Creator
			prefObj.UpdatedBy=creator
			updatedPreferencesJson,err:=json.Marshal(prefObj)
			if err!=nil{
				_preferencesLogger.Errorf("setPreferences : Marshalling Error : " + string(err.Error()))
				replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
				errorData="Marshalling Error :"+replaceErr
				jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
				return shim.Error(jsonResp)
			}
			if isValid,errMsg:=isValidPreferences(prefObj);!isValid{
				return shim.Error(errMsg)
			}
			err=stub.PutState(prefObj.Phone,updatedPreferencesJson)
			if err!=nil{
				_preferencesLogger.Errorf("setPreferences:PutState is Failed :"+string(err.Error()))
				jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\"Unable to update the Preferences\"}"
				return shim.Error(jsonResp)
			}
			_preferencesLogger.Infof("setPreferences:Prefernces updated successfull for msisdn is :"+string(prefObj.Phone))
			err=stub.SetEvent(_UpdateEvent,[]byte(string(updatedPreferencesJson)))
			if err!=nil{
				_preferencesLogger.Errorf("updatePreferences:Event Not Generated for Event is:"+string(_UpdateEvent))
				replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
				errorData="Event is Not Generated "+replaceErr
				jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
				return shim.Error(jsonResp)
			}
			_preferencesLogger.Infof("setPreferences:EventPayload is :"+string(updatedPreferencesJson))
		}else{
			_preferencesLogger.Errorf("setPreferences:Unauthorized Operator is trying to update Existing Preferences")
			jsonResp="{\"Data\":"+prefObj.Phone+",\"ErrorDetails\":\"Access Denied for Unknown Operator\"}"
			return shim.Error(jsonResp)
		}
	}
	resultData:=map[string]interface{}{
		"trxnid":stub.GetTxID(),
		"msisdn":prefObj.Phone,
		"message":"Add Preferences Success",
	}
	respJson,_:=json.Marshal(resultData)
	return shim.Success(respJson)
}



//==============================================================
//portOut for Ownership transfer from Donor to acceptor
//==============================================================
func  (pm *PreferencesManager) portOut(stub shim.ChaincodeStubInterface,args []string)pb.Response{
        if len(args) != 1 {
                _preferencesLogger.Errorf("portOut : Invalid  Number Of Arguments provided for transaction .")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
	var portOutObj Preference
        err:=json.Unmarshal([]byte(args[0]),&portOutObj)
        if err!=nil{
                errorKey=args[0]
                errorData="Invalid json provided as input"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("portOut:"+string(jsonResp))
                return shim.Error(jsonResp)
        }
        _,creator:=pm.getInvokerIdentity(stub)
        preferencesExist,err:=stub.GetState(portOutObj.Phone)
        if err!=nil{
                errorKey=portOutObj.Phone
		replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                errorData="GetState is Failed :"+replaceErr
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("updatePreferences:"+string(jsonResp))
                return shim.Error(jsonResp)
        }

	 if preferencesExist==nil{
                _preferencesLogger.Error("portOut:Preferences Not Exists for msisdn  :"+string(portOutObj.Phone))
		jsonResp="{\"Data\":"+portOutObj.Phone+",\"ErrorDetails\":\"No Existing Preferences\"}"
                return shim.Error(jsonResp)
        }else{
                var updatedBy string
                preference:=Preference{}
                err:=json.Unmarshal(preferencesExist,&preference)
                if err!=nil{
                        _preferencesLogger.Errorf("portOut:Existing PreferenceData unmarshalling Error :"+string(err.Error()))
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Unmarshalling Error :"+replaceErr
			jsonResp="{\"Data\":"+portOutObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
                }
                updatedBy=preference.UpdatedBy
                if strings.Compare(updatedBy,creator)==0{
		        preference.ServiceProvider=portOutObj.ServiceProvider
			preference.UpdateTs=portOutObj.UpdateTs
			preference.Lrn=portOutObj.Lrn
			preference.ServiceAreaCode=portOutObj.ServiceAreaCode
			uby:=dltDomainNames[portOutObj.ServiceProvider]
			if uby==""{
				_preferencesLogger.Errorf("portOut: Invalid ServiceProvider :"+string(portOutObj.ServiceProvider))
				jsonResp="{\"Data\":"+portOutObj.ServiceProvider+",\"ErrorDetails\":\"Invalid ServiceProvider\"}"
				return shim.Error(jsonResp)
			}else{
	                        preference.UpdatedBy=uby
			}
                        portOutJson,err:=json.Marshal(preference)
                        if err!=nil{
                                _preferencesLogger.Errorf("portOut: Marshalling Error : " + string(err.Error()))
				replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
				errorData="Marshalling Error :"+replaceErr
				jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
				return shim.Error(jsonResp)
                        }
                        if isValid,errMsg:=isValidParameters(preference);!isValid{
                                return shim.Error(errMsg)
                        }
                        err=stub.PutState(preference.Phone,portOutJson)
                        if err!=nil{
                                _preferencesLogger.Errorf("portOut:PutState is Failed :"+string(err.Error()))
				jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\"Unable to PortOut  the Preferences\"}"
				return shim.Error(jsonResp)
                        }
			_preferencesLogger.Infof("portOut:Preferences Portout is succesfull for Msisdn is :"+string(preference.Phone))
                        err=stub.SetEvent(_PortOutEvent,[]byte(string(portOutJson)))
                        if err!=nil{
                                _preferencesLogger.Errorf("portOut:Event Not Generated for Event is:"+string(_PortOutEvent))
				replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
				errorData="Event is Not Generated "+replaceErr
				jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
				return shim.Error(jsonResp)
                        }
			_preferencesLogger.Infof("portOut:EventPayload is :"+string(portOutJson))
                        resultData:=map[string]interface{}{
                                "trxnid":stub.GetTxID(),
                                "msisdn":preference.Phone,
                                "message":"Portout is Success",
                        }
                        respJson,_:=json.Marshal(resultData)
                        return shim.Success(respJson)
                }else{
                        _preferencesLogger.Errorf("portOut:Unauthorized Operator is trying to portOut")
			jsonResp="{\"Data\":"+portOutObj.Phone+",\"ErrorDetails\":\"Access Denied for Unknown Operator\"}"
                        return shim.Error(jsonResp)
                }
        }
}



//=================================================================================================================
//deletePreferences for Removing or to churn out preference from DL based on MSISDN on successful certificate check
//=================================================================================================================
func (pm *PreferencesManager) deletePreferences(stub shim.ChaincodeStubInterface,args []string)pb.Response{
         if len(args)!=1{
                _preferencesLogger.Errorf("deletePreferences : Incorrect Number Of Arguments, Excepted <msisdn>.")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
	var updateStatusObj Preference
	err:=json.Unmarshal([]byte(args[0]),&updateStatusObj)
	if err!=nil{
                errorKey=args[0]
                errorData="Invalid json provided as input"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("deletePreferences:"+string(jsonResp))
                return shim.Error(jsonResp)
        }
	_,creator:=pm.getInvokerIdentity(stub)
        preferencesExist, err := stub.GetState(updateStatusObj.Phone)
        if err != nil {
                errorKey=updateStatusObj.Phone
		replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                errorData="GetState is Failed :"+replaceErr
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("deletePreferences:"+string(jsonResp))
                return shim.Error(jsonResp)
        }
        if preferencesExist==nil{
                _preferencesLogger.Error("deletePreferences:Preferences Not Exists for msisdn  :"+string(updateStatusObj.Phone))
		jsonResp="{\"Data\":"+updateStatusObj.Phone+",\"ErrorDetails\":\"No Existing Preferences\"}"
                return shim.Error(jsonResp)
        }else{

                var updatedBy string
                preference:=Preference{}
                err:=json.Unmarshal(preferencesExist,&preference)
                if err!=nil{
                        _preferencesLogger.Errorf("deletePreferences:Existing PreferenceData unmarshalling Error :"+string(err.Error()))
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Unmarshalling Error :"+replaceErr
			jsonResp="{\"Data\":"+updateStatusObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)

                }
                updatedBy=preference.UpdatedBy
                if strings.Compare(updatedBy,creator)==0{
			preference.UpdateTs=updateStatusObj.UpdateTs
			preference.UpdatedBy=creator
			preference.Status="T"
			updateStatusJson,err:=json.Marshal(preference)
                        if err!=nil{
                                _preferencesLogger.Errorf("deletePreferences:Marshalling Error :"+string(err.Error()))
				replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
				errorData="Marshalling Error :"+replaceErr
				jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
				return shim.Error(jsonResp)
                        }
			err=stub.PutState(preference.Phone,updateStatusJson)
			if err!=nil{
				_preferencesLogger.Errorf("deletePreferences:PutState is Failed:"+string(err.Error()))
				jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\"Unable to Delete the Preferences\"}"
				return shim.Error(jsonResp)
			}
			_preferencesLogger.Infof("deletePreferences :Preferences Deleted successfull for Msisdn is :"+string(preference.Phone))
			evtpayload:="{\"msisdn\":"+preference.Phone+",\"status\":\""+preference.Status+"\",\"uts\":\""+preference.UpdateTs+"\"}"
                        err=stub.SetEvent(_DeleteEvent,[]byte(evtpayload))
                        if err!=nil{
                                _preferencesLogger.Errorf("deletePreferences:Event Not Generated For Event "+string(_DeleteEvent))
				replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
				errorData="Event is Not Generated "+replaceErr
				jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
				return shim.Error(jsonResp)
                        }
			_preferencesLogger.Infof("deletePreferences:EventPayload is :"+string(evtpayload))

                        resultData:=map[string]interface{}{
                                "trxnid":stub.GetTxID(),
                                "msisdn":preference.Phone,
                                "message":"Delete Preferences Success",
                        }
                        respJson,_:=json.Marshal(resultData)
                        return shim.Success(respJson)

                }else{
                        _preferencesLogger.Errorf("deletePreferences:Unauthorized Operator is trying to Delete  Preferences")
			jsonResp="{\"Data\":"+updateStatusObj.Phone+",\"ErrorDetails\":\"Access Denied for Unknown Operator\"}"
                        return shim.Error(jsonResp)
                }
        }
}


//==============================================================================================
//snapBackChurn  for Ownership transfer from acceptor to donor 
//==============================================================================================
func(pm *PreferencesManager) snapBackChurn(stub shim.ChaincodeStubInterface,args []string)pb.Response{
	if len(args)!=1{
		_preferencesLogger.Errorf("snapBackChurn:Invalid Number of arguments are provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
	}
	var snapBackObj Preference
        err:=json.Unmarshal([]byte(args[0]),&snapBackObj)
        if err!=nil{
                errorKey=args[0]
                errorData="Invalid json provided as input"
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("portOut:"+string(jsonResp))
                return shim.Error(jsonResp)
        }
        _,creator:=pm.getInvokerIdentity(stub)
        preferencesExist,err:=stub.GetState(snapBackObj.Phone)
        if err!=nil{
                errorKey=snapBackObj.Phone
                replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                errorData="GetState is Failed :"+replaceErr
                jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("updatePreferences:"+string(jsonResp))
                return shim.Error(jsonResp)
        }

         if preferencesExist==nil{
                _preferencesLogger.Error("snapBackChrun:Preferences Not Exists for msisdn  :"+string(snapBackObj.Phone))
                jsonResp="{\"Data\":"+snapBackObj.Phone+",\"ErrorDetails\":\"No Existing Preferences\"}"
                return shim.Error(jsonResp)
        }else{
                var updatedBy string
                preference:=Preference{}
                err:=json.Unmarshal(preferencesExist,&preference)
                if err!=nil{
                        _preferencesLogger.Errorf("snapBackChurn:Existing PreferenceData unmarshalling Error :"+string(err.Error()))
                        replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                        errorData="Unmarshalling Error :"+replaceErr
                        jsonResp="{\"Data\":"+snapBackObj.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
                }
                updatedBy=preference.UpdatedBy
                if strings.Compare(updatedBy,creator)==0{
                        preference.ServiceProvider=snapBackObj.ServiceProvider
                        preference.UpdateTs=snapBackObj.UpdateTs
                        preference.Lrn=snapBackObj.Lrn
			preference.Status="T"
                        preference.ServiceAreaCode=snapBackObj.ServiceAreaCode
                        uby:=dltDomainNames[snapBackObj.ServiceProvider]
                        if uby==""{
                                _preferencesLogger.Errorf("snapBackChurn: Invalid ServiceProvider :"+string(snapBackObj.ServiceProvider))
                                jsonResp="{\"Data\":"+snapBackObj.ServiceProvider+",\"ErrorDetails\":\"Invalid ServiceProvider\"}"
                                return shim.Error(jsonResp)
                        }else{
                                preference.UpdatedBy=uby
                        }
                        snapBackJson,err:=json.Marshal(preference)
                        if err!=nil{
                                _preferencesLogger.Errorf("snapBackChurn: Marshalling Error : " + string(err.Error()))
                                replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                                errorData="Marshalling Error :"+replaceErr
                                jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                                return shim.Error(jsonResp)
                        }
                        if isValid,errMsg:=isValidParameters(preference);!isValid{
                                return shim.Error(errMsg)
                        }
			  err=stub.PutState(preference.Phone,snapBackJson)
                        if err!=nil{
                                _preferencesLogger.Errorf("snapBackChurn:PutState is Failed :"+string(err.Error()))
                                jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\"Unable to PortOut  the Preferences\"}"
                                return shim.Error(jsonResp)
                        }
                        _preferencesLogger.Infof("snapBackChurn:snapBackChurn is succesfull for Msisdn is :"+string(preference.Phone))
                        err=stub.SetEvent(_SnapBackEvent,[]byte(string(snapBackJson)))
                        if err!=nil{
                                _preferencesLogger.Errorf("snapBackChurn:Event Not Generated for Event is:"+string(_SnapBackEvent))
                                replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                                errorData="Event is Not Generated "+replaceErr
                                jsonResp="{\"Data\":"+preference.Phone+",\"ErrorDetails\":\""+errorData+"\"}"
                                return shim.Error(jsonResp)
                        }
                        _preferencesLogger.Infof("snapBackChurn:EventPayload is :"+string(snapBackJson))
                        resultData:=map[string]interface{}{
                                "trxnid":stub.GetTxID(),
                                "msisdn":preference.Phone,
                                "message":"SnapBackChurn is Success",
                        }
                        respJson,_:=json.Marshal(resultData)
                        return shim.Success(respJson)
                }else{
                        _preferencesLogger.Errorf("snapBackChurn:Unauthorized Operator is trying to snapBack")
                        jsonResp="{\"Data\":"+snapBackObj.Phone+",\"ErrorDetails\":\"Access Denied for Unknown Operator\"}"
                        return shim.Error(jsonResp)
                }
        }
}



//========================================================
//batchPreferences for Uploading Bulk Preferences into DL
//========================================================
func  (pm *PreferencesManager) batchPreferences(stub shim.ChaincodeStubInterface,args []string)pb.Response{
        if len(args)==0{
                _preferencesLogger.Errorf("batchPreferences:Invalid Number of arguments provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
        var msisdn_f []string
        for i:=0;i<len(args);i++{
		var prefObj Preference
                err:=json.Unmarshal([]byte(args[i]),&prefObj)
                if err!=nil{
                        errorKey=args[i]
                        errorData="Invalid json provided as input"
                        jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                        _preferencesLogger.Error("batchPreferences:"+string(jsonResp))
                        msisdn_f=append(msisdn_f,prefObj.Phone)
                        continue
                }
                _,creator:=pm.getInvokerIdentity(stub)
                preferencesExist,err:=stub.GetState(prefObj.Phone)
                if err!=nil{
                        errorKey=prefObj.Phone
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                        errorData="GetState is Failed :"+replaceErr
                        jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                        _preferencesLogger.Error("batchPreferences:"+string(jsonResp))
                        msisdn_f=append(msisdn_f,prefObj.Phone)
                        continue

                }
		 if preferencesExist==nil{
                        prefObj.ObjType="Preferences"
			prefObj.Creator=creator
                        prefObj.UpdatedBy=creator
                        preferencesJson,err:=json.Marshal(prefObj)
                        if err!=nil{
                                _preferencesLogger.Errorf("batchPreferences : Marshalling Error : " + string(err.Error()))
                                msisdn_f=append(msisdn_f,prefObj.Phone)
                                continue
                        }
                        if isValid,errMsg:=isValidPreferences(prefObj);!isValid{
				_preferencesLogger.Errorf("batchPreferences:"+string(errMsg))
                                msisdn_f=append(msisdn_f,string(prefObj.Phone))
                                continue
                        }
                        err=stub.PutState(prefObj.Phone,preferencesJson)
                        if err!=nil{
                                _preferencesLogger.Errorf("batchPreferences:PutState is Failed :"+string(err.Error()))
                                msisdn_f=append(msisdn_f,string(prefObj.Phone))
                                continue
                        }
			_preferencesLogger.Infof("batchPreferences:Preferences added successfull for Msisdn is :"+string(prefObj.Phone))
                        err=stub.SetEvent(_AddEvent,[]byte(string(preferencesJson)))
                        if err!=nil{
                                _preferencesLogger.Errorf("batchPreferences:Event Not Generated for Event is:"+string(_AddEvent))
                                msisdn_f=append(msisdn_f,string(prefObj.Phone))
                                continue
                        }
			_preferencesLogger.Infof("batchPreferences:Event Payload is :"+string(preferencesJson))
                }else{
                        var updatedBy string
                        preference:=Preference{}
                        err:=json.Unmarshal(preferencesExist,&preference)
                        if err!=nil{
                                _preferencesLogger.Errorf("batchPreferences:Existing PreferenceData unmarshalling Error :"+string(err.Error()))
                                msisdn_f=append(msisdn_f,prefObj.Phone)
                                continue
                        }
                        updatedBy=preference.UpdatedBy
			if strings.Compare(updatedBy,creator)==0{
				prefObj.ObjType="Preferences"
				prefObj.CreateTs=preference.CreateTs
				prefObj.Creator=preference.Creator
				prefObj.UpdatedBy=creator
				updatedPreferencesJson,err:=json.Marshal(prefObj)
				if err!=nil{
					_preferencesLogger.Errorf("batchPreferences : Marshalling Error : " + string(err.Error()))
					msisdn_f=append(msisdn_f,prefObj.Phone)
					continue
				}
				if isValid,errMsg:=isValidPreferences(prefObj);!isValid{
					_preferencesLogger.Errorf("batchPreferences:"+string(errMsg))
					msisdn_f=append(msisdn_f,prefObj.Phone)
					continue
				}
				err=stub.PutState(prefObj.Phone,updatedPreferencesJson)
				if err!=nil{
					_preferencesLogger.Errorf("batchPreferences:PutState is Failed :"+string(err.Error()))
					msisdn_f=append(msisdn_f,prefObj.Phone)
					continue
				}
				_preferencesLogger.Infof("batchPreferences:Preferences updated success full for msisdn is :"+string(prefObj.Phone))
				err=stub.SetEvent(_UpdateEvent,[]byte(string(updatedPreferencesJson)))
				if err!=nil{
					_preferencesLogger.Errorf("batchPreferences:Event Not Generated for Event is:"+string(_UpdateEvent))
					msisdn_f=append(msisdn_f,prefObj.Phone)
					continue
				}
				_preferencesLogger.Infof("batchPreferences:Event Payload is :"+string(updatedPreferencesJson))

			}else{
				 _preferencesLogger.Errorf("batchPreferences:Unauthorized Operator is trying to update Existing Preferences")
				 msisdn_f=append(msisdn_f,prefObj.Phone)
				 continue
			}
		}
	}
        resultData:=map[string]interface{}{
                "trxnid":stub.GetTxID(),
                "msisdn_f":msisdn_f,
                "message":"Batch Preferences Success",
        }
        respJson,_:=json.Marshal(resultData)
        return shim.Success(respJson)
}


//=======================================================================
//batchPortOut for Ownership transfer from Donor to acceptor of BulkData
//======================================================================
func  (pm *PreferencesManager) batchPortOut(stub shim.ChaincodeStubInterface,args []string)pb.Response{
        if len(args)==0{
                _preferencesLogger.Errorf("batchPortOut:Invalid Number of arguments provided for transaction.")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
        var msisdn_f []string
        for i:=0;i<len(args);i++{
		var portOutObj Preference
                err:=json.Unmarshal([]byte(args[i]),&portOutObj)
                if err!=nil{
                        errorKey=args[i]
                        errorData="Invalid json provided as input"
                        jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                        _preferencesLogger.Error("batchPortOut:"+string(jsonResp))
                        msisdn_f=append(msisdn_f,portOutObj.Phone)
                        continue
                }
                _,creator:=pm.getInvokerIdentity(stub)
                preferencesExist,err:=stub.GetState(portOutObj.Phone)
                if err!=nil{
                        errorKey=portOutObj.Phone
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                        errorData="GetState is Failed :"+replaceErr
                        jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                        _preferencesLogger.Error("batchPortOut:"+string(jsonResp))
                        msisdn_f=append(msisdn_f,portOutObj.Phone)
                        continue
                }
		if preferencesExist==nil{
                        _preferencesLogger.Error("batchPortOut:Preferences Not Exists for msisdn  :"+string(portOutObj.Phone))
                        continue
                }else{
                        var updatedBy string
                        preference:=Preference{}
                        err:=json.Unmarshal(preferencesExist,&preference)
                        if err!=nil{
                                _preferencesLogger.Errorf("batchPortOut:Existing PreferenceData unmarshalling Error :"+string(err.Error()))
                                msisdn_f=append(msisdn_f,portOutObj.Phone)
                                continue
                        }
                        updatedBy=preference.UpdatedBy
                        if strings.Compare(updatedBy,creator)==0{
				preference.ServiceProvider=portOutObj.ServiceProvider
				preference.UpdateTs=portOutObj.UpdateTs
				preference.Lrn=portOutObj.Lrn
				preference.ServiceAreaCode=portOutObj.ServiceAreaCode
				uby:=dltDomainNames[portOutObj.ServiceProvider]
				if uby==""{
					_preferencesLogger.Errorf("batchPortOut:Invalid Service Provider :"+string(portOutObj.ServiceProvider))
					msisdn_f=append(msisdn_f,portOutObj.Phone)
					continue
				}else{
	                                preference.UpdatedBy = uby
				}
                                portOutJson,err:=json.Marshal(preference)
                                if err!=nil{
                                        _preferencesLogger.Errorf("batchPortOut: Marshalling Error : " + string(err.Error()))
                                        msisdn_f=append(msisdn_f,portOutObj.Phone)
                                        continue
                                }
                                if isValid,errMsg:=isValidParameters(preference);!isValid{
					_preferencesLogger.Errorf("batchPortOut:"+string(errMsg))
                                        msisdn_f=append(msisdn_f,portOutObj.Phone)
                                        continue
                                }
                                err=stub.PutState(preference.Phone,portOutJson)
                                if err!=nil{
                                        _preferencesLogger.Errorf("batchPortOut:PutState is Failed :"+string(err.Error()))
                                        msisdn_f=append(msisdn_f,preference.Phone)
                                        continue
                                }
				_preferencesLogger.Infof("batchPortout:portout is successfull for Msisdn is :"+string(preference.Phone))
                                err=stub.SetEvent(_PortOutEvent,[]byte(string(portOutJson)))
                                if err!=nil{
                                        _preferencesLogger.Errorf("batchPortOut:Event Not Generated for Event is:"+string(_PortOutEvent))
                                        msisdn_f=append(msisdn_f,preference.Phone)
                                        continue
                                }
				_preferencesLogger.Infof("batchPortOut:EventPayload is :"+string(portOutJson))

			}else{
				_preferencesLogger.Errorf("batchPortOut:Unauthorized Operator is trying to portOut")
				msisdn_f=append(msisdn_f,preference.Phone)
				continue
			}
		}
	}
	resultData:=map[string]interface{}{
                "trxnid":stub.GetTxID(),
                "msisdn_f":msisdn_f,
                "message":"Batch PortOut Success",
        }
        respJson,_:=json.Marshal(resultData)
        return shim.Success(respJson)
}


//===============================================================================================================================
//batchDeletePreferences for Removing or to churn out preference from DL based on MSISDN on successful certificate check BulkData
//===============================================================================================================================
func  (pm *PreferencesManager) batchDeletePreferences(stub shim.ChaincodeStubInterface,args []string)pb.Response{
        if len(args) == 0 {
                _preferencesLogger.Errorf("bathcDeletePreferences : Invalid Number of arguments are provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
        var msisdn_f []string
	for i:=0;i<len(args);i++{
		var updateStatusObj Preference
		err:=json.Unmarshal([]byte(args[i]),&updateStatusObj)
		if err!=nil{
			errorKey=args[0]
			errorData="Invalid json provided as input"
			jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
			_preferencesLogger.Error("batchDeletePreferences:"+string(jsonResp))
			msisdn_f=append(msisdn_f,updateStatusObj.Phone)
			continue
		}
		_,creator:=pm.getInvokerIdentity(stub)
		preferencesExist, err := stub.GetState(updateStatusObj.Phone)
		if err != nil {
			errorKey=updateStatusObj.Phone
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="GetState is Failed :"+replaceErr
			jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
			_preferencesLogger.Error("batchDeletePreferences:"+string(jsonResp))
			msisdn_f=append(msisdn_f,updateStatusObj.Phone)
			continue
		}
                if preferencesExist==nil{
                        _preferencesLogger.Error("batchDeletePreferences:Preferences Not Exists for msisdn  :"+string(updateStatusObj.Phone))
                        continue
                }else{
                        var updatedBy string
                        preference:=Preference{}
                        err:=json.Unmarshal(preferencesExist,&preference)
                        if err!=nil{
                                _preferencesLogger.Errorf("batchDelete:Existing PreferenceData unmarshalling Error :"+string(err.Error()))
                                msisdn_f=append(msisdn_f,updateStatusObj.Phone)
                                continue
                        }
                        updatedBy=preference.UpdatedBy
                        if strings.Compare(updatedBy,creator)==0{
				preference.UpdateTs=updateStatusObj.UpdateTs
				preference.UpdatedBy=creator
				preference.Status="T"
				updateStatusJson,err:=json.Marshal(preference)
				if err!=nil{
					_preferencesLogger.Errorf("batchDeletePreferences:Marshalling Error :"+string(err.Error()))
					msisdn_f=append(msisdn_f,preference.Phone)
					continue
				}
				err=stub.PutState(preference.Phone,updateStatusJson)
				if err!=nil{
					_preferencesLogger.Errorf("batchDeletePreferences:PutState is Failed:"+string(err.Error()))
					msisdn_f=append(msisdn_f,preference.Phone)
					continue
				}
				_preferencesLogger.Infof("batchDeletePreferences :Preferences Deleted successfull for Msisdn is :"+string(preference.Phone))
				evtpayload:="{\"msisdn\":\""+preference.Phone+"\",\"status\":\""+preference.Status+"\",\"uts\":\""+preference.UpdateTs+"\"}"
				err=stub.SetEvent(_DeleteEvent,[]byte(evtpayload))
				if err!=nil{
					_preferencesLogger.Errorf("batchDeletePreferences:Event Not Generated For Event "+string(_DeleteEvent))
					msisdn_f=append(msisdn_f,preference.Phone)
					continue
				}
				_preferencesLogger.Infof("batchDeletePreferences:EventPayload is :"+string(evtpayload))
                        }else{
                                _preferencesLogger.Errorf("batchDeletePreferences:Unauthorized Operator is trying to Delete  Preferences")
                                msisdn_f=append(msisdn_f,preference.Phone)
				continue
                        }
                }
        }
          resultData:=map[string]interface{}{
                "trxnid":stub.GetTxID(),
                "msisdn_f":msisdn_f,
                "message":"Batch Delete Success",
        }
        respJson,_:=json.Marshal(resultData)
        return shim.Success(respJson)
}


//====================================================================================================
//batchSnapBackChurn for Ownership transfer from acceptor to donor in BULK 
//====================================================================================================
func(pm *PreferencesManager) batchSnapBackChurn(stub shim.ChaincodeStubInterface,args []string) pb.Response{
	 if len(args) == 0 {
                _preferencesLogger.Errorf("batchSnapBackChurn : Invalid Number of arguments are provided for transaction")
                jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
        var msisdn_f []string
	for i:=0;i<len(args);i++{
                var snapBackObj Preference
                err:=json.Unmarshal([]byte(args[i]),&snapBackObj)
                if err!=nil{
                        errorKey=args[i]
                        errorData="Invalid json provided as input"
                        jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                        _preferencesLogger.Error("batchSnapBackChurn:"+string(jsonResp))
                        msisdn_f=append(msisdn_f,snapBackObj.Phone)
                        continue
                }
                _,creator:=pm.getInvokerIdentity(stub)
                preferencesExist,err:=stub.GetState(snapBackObj.Phone)
                if err!=nil{
                        errorKey=snapBackObj.Phone
                        replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                        errorData="GetState is Failed :"+replaceErr
                        jsonResp="{\"Data\":"+errorKey+",\"ErrorDetails\":\""+errorData+"\"}"
                        _preferencesLogger.Error("batchPortOut:"+string(jsonResp))
                        msisdn_f=append(msisdn_f,snapBackObj.Phone)
                        continue
                }
                if preferencesExist==nil{
                        _preferencesLogger.Error("batchSnapBackChurn:Preferences Not Exists for msisdn  :"+string(snapBackObj.Phone))
                        continue
                }else{
                        var updatedBy string
                        preference:=Preference{}
                        err:=json.Unmarshal(preferencesExist,&preference)
                        if err!=nil{
                                _preferencesLogger.Errorf("batchSanpBackChurn:Existing PreferenceData unmarshalling Error :"+string(err.Error()))
                                msisdn_f=append(msisdn_f,snapBackObj.Phone)
                                continue
                        }
                        updatedBy=preference.UpdatedBy
                        if strings.Compare(updatedBy,creator)==0{
                                preference.ServiceProvider=snapBackObj.ServiceProvider
                                preference.UpdateTs=snapBackObj.UpdateTs
                                preference.Lrn=snapBackObj.Lrn
                                preference.Status="T"
                                preference.ServiceAreaCode=snapBackObj.ServiceAreaCode
                                uby:=dltDomainNames[snapBackObj.ServiceProvider]
                                if uby==""{
                                        _preferencesLogger.Errorf("batchSnapBackChurn:Invalid Service Provider :"+string(snapBackObj.ServiceProvider))
                                        msisdn_f=append(msisdn_f,snapBackObj.Phone)
                                        continue
                                }else{
                                        preference.UpdatedBy = uby
                                }
                                snapBackJson,err:=json.Marshal(preference)
                                if err!=nil{
                                        _preferencesLogger.Errorf("batchSnapBackChurn: Marshalling Error : " + string(err.Error()))
                                        msisdn_f=append(msisdn_f,snapBackObj.Phone)
                                        continue
                                }
                                if isValid,errMsg:=isValidParameters(preference);!isValid{
                                        _preferencesLogger.Errorf("batchSnapBackChurn:"+string(errMsg))
                                        msisdn_f=append(msisdn_f,snapBackObj.Phone)
                                        continue
                                }
				err=stub.PutState(preference.Phone,snapBackJson)
                                if err!=nil{
                                        _preferencesLogger.Errorf("batchSnapBackChurn:PutState is Failed :"+string(err.Error()))
                                        msisdn_f=append(msisdn_f,preference.Phone)
                                        continue
                                }
                                _preferencesLogger.Infof("batchSnapBackChurn:SnapBack is successfull for Msisdn is :"+string(preference.Phone))
                                err=stub.SetEvent(_SnapBackEvent,[]byte(string(snapBackJson)))
                                if err!=nil{
                                        _preferencesLogger.Errorf("batchSnapBackChurn:Event Not Generated for Event is:"+string(_SnapBackEvent))
                                        msisdn_f=append(msisdn_f,preference.Phone)
                                        continue
                                }
                                _preferencesLogger.Infof("batchSnapBackChurn:EventPayload is :"+string(snapBackJson))

                        }else{
                                _preferencesLogger.Errorf("batchSnapBackChurn:Unauthorized Operator is trying to portOut")
                                msisdn_f=append(msisdn_f,preference.Phone)
                                continue
                        }
                }
        }
        resultData:=map[string]interface{}{
                "trxnid":stub.GetTxID(),
                "msisdn_f":msisdn_f,
                "message":"Batch SnapBackChurn Success",
        }
        respJson,_:=json.Marshal(resultData)
        return shim.Success(respJson)
}


//========================================================================================================
//getPreferencesByMsisdn: get PreferencesData from DL based on Msisdn
//=========================================================================================================
func (pm *PreferencesManager) getPreferencesByMsisdn(stub shim.ChaincodeStubInterface,args []string) pb.Response{
	if len(args)!=1{
		_preferencesLogger.Errorf("getPreferencesByMsisdn:Invalid Number of arguments are provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
	}
	var records []Preference
	preferencesExist, err := stub.GetState(args[0])
	if err!=nil{
		errorKey=args[0]
		replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
                errorData="GetState is Failed :"+replaceErr
                jsonResp="{\"Data\":\""+errorKey+"\",\"ErrorDetails\":\""+errorData+"\"}"
                _preferencesLogger.Error("batchDeletePreferences:"+string(jsonResp))
		return shim.Error(string(jsonResp))
	}
	if preferencesExist==nil{
		_preferencesLogger.Errorf("getPreferencesByMsisdn:No Existing preferences for Msisdn:"+string(args[0]))
		jsonResp="{\"Data\":\""+args[0]+"\",\"ErrorDetails\":\"No Existing Preferences\"}"
                return shim.Error(jsonResp)
	}else{
		preference:=Preference{}
		err:=json.Unmarshal(preferencesExist,&preference)
		if err!=nil{
			_preferencesLogger.Errorf("getPreferencesByMsisdn::Existing PreferenceData unmarshalling Error"+string(err.Error()))
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Unmarshalling Error :"+replaceErr
			jsonResp="{\"Data\":\""+args[0]+"\",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
		}
		records=append(records,preference)
		resultData:=map[string]interface{}{
			"status":"true",
			"preferences":records[0],
		}
		respJson,_:=json.Marshal(resultData)
		return shim.Success(respJson)
	}
}

//======================================================================================
//queryPreferences RichQuery for Obtaining Preference prefObj
//======================================================================================
func  (pm *PreferencesManager) queryPreferences(stub shim.ChaincodeStubInterface,args []string) pb.Response{
	if len(args)!=1{
		_preferencesLogger.Errorf("queryPreferences:Invalid number of arguments are provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
	}
	var records []Preference
	queryString:=args[0]
        _preferencesLogger.Infof("Query Selector : "+string(queryString))
        resultsIterator,err:=stub.GetQueryResult(queryString)
        if err!=nil{
                _preferencesLogger.Error("queryPreferences:GetQueryResult is Failed with error :"+string(err.Error()))
		errorData="GetQueryResult Error :"+string(err.Error())
		jsonResp="{\"Data\":"+args[0]+",\"ErrorDetails\":\""+errorData+"\"}"
                return shim.Error(jsonResp)
        }
        for resultsIterator.HasNext(){
		record:=Preference{}
                recordBytes,_:=resultsIterator.Next()
		if (string(recordBytes.Value))==""{
			continue
		}
		err=json.Unmarshal(recordBytes.Value,&record)
		if err!=nil{
			_preferencesLogger.Errorf("queryPreferences:Unable to unmarshal Preferences retrieved :"+string(err.Error()))
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Unmarshalling Error :"+replaceErr
			jsonResp="{\"Data\":"+string(recordBytes.Value)+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
		}
		records=append(records,record)
	}
	resultData:=map[string]interface{}{
		"status":"true",
		"preferences":records,
	}
	respJson,_:=json.Marshal(resultData)
	return shim.Success(respJson)
}



//========================================================================================
//getHistoryQuery for Getting all history data for msisdn
//=======================================================================================-

func  (pm *PreferencesManager) getHistoryPreferences(stub shim.ChaincodeStubInterface,args []string)pb.Response{
	if len(args)!=1{
		_preferencesLogger.Errorf("getHistoryPreferences:Invalid number of arguments are provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
	var records []Preference
	resultsIterator,err:=stub.GetHistoryForKey(args[0])
	if err!=nil{
                _preferencesLogger.Errorf("getHistoryPreferences:GetHistoryForKey is Failed"+string(err.Error()))
		replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
		errorData="GetHistoryForKey Error :"+replaceErr
		jsonResp="{\"Data\":"+args[0]+",\"ErrorDetails\":\""+errorData+"\"}"
                return shim.Error(jsonResp)
        }
	for resultsIterator.HasNext(){
		record:=Preference{}
		recordBytes,_:=resultsIterator.Next()
		if string(recordBytes.Value)==""{
			continue
		}
		err:=json.Unmarshal(recordBytes.Value,&record)
		if err!=nil{
			_preferencesLogger.Errorf("getHistoryPreferences:Unable to unmarshal Preferences retrieved :"+string(err.Error()))	
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Unmarshalling Error :"+replaceErr
			jsonResp="{\"Data\":"+string(recordBytes.Value)+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
		}
		records=append(records,record)
	}
	resultData:=map[string]interface{}{
		"status":"true",
		"preferences":records,
	}
	respJson,_:=json.Marshal(resultData)
	return shim.Success(respJson)
}


// ===== Example: Pagination with Ad hoc Rich Query ========================================================
// queryPreferencesWithPagination uses a query string, page size and a bookmark to perform a query
// for Preferences. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// =========================================================================================
func (pm *PreferencesManager) queryPreferencesWithPagination(stub shim.ChaincodeStubInterface,args []string)pb.Response{
        if len(args)!=3{
                _preferencesLogger.Errorf("queryPreferencesWithPagination:Invalid number of arguments provided for transaction")
		jsonResp="{\"Data\":"+strconv.Itoa(len(args))+",\"ErrorDetails\":\"Invalid Number of argumnets provided for transaction\"}"
                return shim.Error(jsonResp)
        }
        var records []Preference
        queryString:=args[0]
        pageSize,err:=strconv.ParseInt(args[1],10,32)
        if err!=nil{
                _preferencesLogger.Errorf("queryPreferencesQithPagination:Error while ParseInt is :"+string(err.Error()))
		replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
		errorData="Pagesize parseint error :"+replaceErr
		jsonResp="{\"Data\":"+args[1]+",\"ErrorDetails\":\""+errorData+"\"}"
                return shim.Error(jsonResp)
        }
        bookmark:=args[2]
        resultsIterator,responseMetaData,err:=stub.GetQueryResultWithPagination(queryString,int32(pageSize),bookmark)
        if err!=nil{
                _preferencesLogger.Errorf("queryPreferenncesWithPagination:GetQueryResultWithPagination is Failed :"+string(err.Error()))
		replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
		errorData="GetQueryResultWithPagination Error :"+replaceErr
		jsonResp="{\"Data\":"+args[0]+",\"ErrorDetails\":\""+errorData+"\"}"
                return shim.Error(jsonResp)
        }
        for resultsIterator.HasNext(){
                record:=Preference{}
                recordBytes,_:=resultsIterator.Next()
                if string(recordBytes.Value)==""{
                        continue
                }
                err:=json.Unmarshal(recordBytes.Value,&record)
                if err!=nil{
                        _preferencesLogger.Errorf("getHistoryPreferences:Unable to unmarshal Preferences retrieved :"+string(err.Error()))
			replaceErr := strings.Replace(err.Error(), "\"", " ", -1)
			errorData="Unmarshalling Error :"+replaceErr
			jsonResp="{\"Data\":"+string(recordBytes.Value)+",\"ErrorDetails\":\""+errorData+"\"}"
                        return shim.Error(jsonResp)
                }
                records=append(records,record)
        }
        resultData:=map[string]interface{}{
                "status":"true",
                "preferences":records,
                "recordscount":responseMetaData.FetchedRecordsCount,
                "bookmark":responseMetaData.Bookmark,
        }
        respJson,_:=json.Marshal(resultData)
        return shim.Success(respJson)


}


// ===================================================================================
//main function for the preference ChainCode
// ===================================================================================
func main() {
        err := shim.Start(new(PreferencesManager))
        _preferencesLogger.SetLevel(shim.LogDebug)
        if err != nil {
                _preferencesLogger.Error("Error Starting PreferencesManager Chaincode is " + string(err.Error()))
        } else {
                _preferencesLogger.Info("Starting PreferencesManager Chaincode")
        }
}



