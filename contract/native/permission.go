package native

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	"github.com/Ankr-network/ankr-chain/context"
	ankrcrypto "github.com/Ankr-network/ankr-chain/crypto"
	"github.com/tendermint/tendermint/libs/log"
)

type Permission struct {
	context context.ContextContract
	log     log.Logger
}

func (p* Permission) AddRole(rType ankrcmm.RoleType, name string, pubKey string, contractAddr string) (uint32, string) {
	regx := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !regx.MatchString(name) {
		return code.CodeTypeRoleNameInvalid, fmt.Sprintf("Permission AddRole, invalid role name:%s", name)
	}

	_, _, _, _, err := p.context.LoadRole(name, 0, false)
	if err == nil {
		return code.CodeTypeRoleExisted, fmt.Sprintf("Permission AddRole, role existed:%s", name)
	}

	if rType == ankrcmm.RoleContract {
		if contractAddr == "" {
			return code.CodeTypeRoleContractAddrBlank, fmt.Sprintf("Permission AddRole, blank contract address for contract role name:%s", name)
		}else {
			//The disposing temp undo
			/*cInfo, _, _, _, err := p.context.LoadContract(contractAddr, 0, false)
			if err != nil || cInfo == nil {
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				return code.CodeTypeContractCantFound, fmt.Sprintf("Permission AddRole, can't load contract %s, %s", contractAddr, errStr)
			}

			if p.context.SenderAddr() != cInfo.Owner {
				return code.CodeTypeRoleInvalidAccount, fmt.Sprintf("Permission AddRole, now contract %s owner address, expected %s, got %s", contractAddr, cInfo.Owner, p.context.SenderAddr())
			}*/
		}
	}else {
		//TBD
		return code.CodeTypeRoleUnSupportedType, fmt.Sprintf("Permission AddRole, not support role type: %d", rType)
	}

	p.context.AddRole(rType, name, pubKey, contractAddr)

	return code.CodeTypeOK, ""
}

func (p *Permission) RoleBindAction(roleName string, contractAddr string, actionName string) (uint32, string) {
	rInfo, _, _, _, err := p.context.LoadRole(roleName, 0, false)

	if err != nil {
		return code.CodeTypeRoleNotExisted, fmt.Sprintf("Permission BindAction, role not existed:%s, err:%s", roleName, err.Error())
	}
	if rInfo == nil {
		return code.CodeTypeRoleNotExisted, fmt.Sprintf("Permission BindAction, role not existed:%s", roleName)
	}

	if contractAddr != "" && rInfo.Type != ankrcmm.RoleContract {
		return code.CodeTypeRoleNotMismatch, fmt.Sprintf("Permission BindAction, role mismatch: expected %d, got %d", ankrcmm.RoleContract, rInfo.Type)
	}

	if contractAddr == "" {
		//The disposing temp undo
		/*cInfo, _, _, _, err := p.context.LoadContract(contractAddr, 0, false)
		if err != nil || cInfo == nil {
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			return code.CodeTypeContractCantFound, fmt.Sprintf("Permission BindAction, can't load contract %s, %s", contractAddr, errStr)
		}

		if p.context.SenderAddr() != cInfo.Owner {
			return code.CodeTypeRoleInvalidAccount, fmt.Sprintf("Permission BindAction, now contract %s owner address, expected %s, got %s", contractAddr, cInfo.Owner, p.context.SenderAddr())
		}*/
	}else {
		//TBD
		return code.CodeTypeRoleUnSupportedType, fmt.Sprintf("Permission BindAction, not support bound role type: %d", ankrcmm.RoleGeneral)
	}

	p.context.AddBoundAction(roleName, contractAddr, actionName)

	return code.CodeTypeOK, ""
}

func (p *Permission) AccountBindRole(accAddr string, roleName string) (uint32, string) {
	rInfo, _, _, _, err := p.context.LoadRole(roleName, 0, false)

	if err != nil {
		return code.CodeTypeRoleNotExisted, fmt.Sprintf("Permission BindAction, role not existed:%s, err:%s", roleName, err.Error())
	}
	if rInfo == nil {
		return code.CodeTypeRoleNotExisted, fmt.Sprintf("Permission BindAction, role not existed:%s", roleName)
	}

	p.context.AddBoundRole(accAddr, roleName)

	return code.CodeTypeOK, ""
}

func (p *Permission) verifySignedData(pubKey string, signedData string, toBeVerifiedInfo string) bool {
	pubKeyEd, err := ankrcrypto.DeserilizePubKey(pubKey)
	if err != nil {
		p.log.Error("Permission verifySignedData err", "err", err)
		return false
	}

	addr := pubKeyEd.Address().String()
	if len(addr) != ankrcmm.KeyAddressLen {
		p.log.Error("Permission verifySignedData invalid pubKey", "err", err)
		return false
	}

	sum := sha256.Sum256([]byte(toBeVerifiedInfo))
	return pubKeyEd.VerifyBytes(sum[:32], []byte(signedData))
}

func (p *Permission) VerifyAuthority(accAddr string, contractAddr string, actionName string, authInfos string, toBeVerifiedInfo string) bool {
	accRoles, err := p.context.LoadBoundRoles(accAddr)
	if err != nil {
		p.log.Error("Permission VerifyAuthority err", "err", err)
		return false
	}

	roleSignMap := make(map[string]string)
	auths := strings.Split(authInfos,",")
	for _, auth := range auths {
		authArr := strings.Split(auth, ":")
		if len(authArr) != 2 {
			p.log.Error("Permission VerifyAuthority, invalid auth infos", "accAddr", accAddr, "contractAddr", contractAddr, "actionName", "authInfos", authInfos)
			return false
		}
		roleName   := authArr[0]
		signedData := authArr[1]

		roleSignMap[roleName] = signedData
	}

	rbaInfoList := p.context.LoadBoundAction(contractAddr, actionName)
	if len(rbaInfoList.RoleBounds) == 0 {
		p.log.Info("Permission VerifyAuthority, there is any bound role")
		return true
	}

	for _, rbaInfo := range rbaInfoList.RoleBounds {
		isHaveRole := false
		for _, accRole := range accRoles {
			if accRole == rbaInfo.Name {
				isHaveRole = true
				break
			}
		}

		if isHaveRole {
			if signedData, ok := roleSignMap[rbaInfo.Name]; ok {
				rInfo, _, _, _, err := p.context.LoadRole(rbaInfo.Name, 0, false)
				if err != nil {
					p.log.Error("Permission VerifyAuthority, can load role", "roleName", rbaInfo.Name, "err", err)
					return false
				}
				isPassVerify := p.verifySignedData(rInfo.PubKey, signedData,toBeVerifiedInfo)
				if !isPassVerify {
					p.log.Error("Permission VerifyAuthority, verifySignedData fail", "pubKey", rInfo.PubKey, "signedData", signedData)
					return false
				}
			}
		}else {
			p.log.Error("Permission VerifyAuthority, account hasn't the responding role", "accAddr", accAddr)
			return false
		}
	}

	return true
}
