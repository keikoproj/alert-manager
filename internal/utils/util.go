package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/pkg/log"
	"strings"
)

//ExportParamsChecksum function calculates checksum if exportParams is not empty
func ExportParamsChecksum(ctx context.Context, exportedParams []string) (bool, string) {
	log := log.Logger(ctx, "internal.utils", "util", "ExportParamsChecksum")
	if len(exportedParams) == 0 {
		return false, ""
	}
	log.V(4).Info("exportedParams are not empty")
	return true, calculateChecksum(ctx, strings.Join(exportedParams, ""))
}

//CalculateAlertConfigChecksum function calculates hash value for Alert Config
func CalculateAlertConfigChecksum(ctx context.Context, input v1alpha1.Config, global v1alpha1.OrderedMap) (bool, string) {
	log := log.Logger(ctx, "internal.utils", "util", "CalculateAlertConfigChecksum")

	globalMap := MergeMaps(ctx, global, input.Params)
	// Now, this should have the params from both global and local to calculate checksum
	input.Params = globalMap
	jsonData, err := json.Marshal(input)
	if err != nil {
		log.Error(err, "Unable to marshal the input request")
		return false, ""
	}
	return true, calculateChecksum(ctx, string(jsonData))
}

//MergeMaps function used to merge two maps i.e, baseParams fields gets overwritten by overwriteParams if exists
func MergeMaps(ctx context.Context, baseParams map[string]string, overwriteParams map[string]string) map[string]string {
	log := log.Logger(ctx, "internal.utils", "util", "MergeMaps")
	log.V(4).Info("merging maps")
	globalMap := make(v1alpha1.OrderedMap)
	// Order must be first load global
	for k, v := range baseParams {
		globalMap[k] = v
	}

	// overwrite if there is anything specified in individual section
	for k, v := range overwriteParams {
		globalMap[k] = v
	}
	return globalMap
}

//CalculateChecksum is an exported function
func CalculateChecksum(ctx context.Context, input string) string {
	return calculateChecksum(ctx, input)
}

//calculateChecksum function calculates checksum for the given string
func calculateChecksum(ctx context.Context, input string) string {
	log := log.Logger(ctx, "internal.utils", "util", "calculateChecksum")
	log.V(4).Info("calculating checksum", "input", input)
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

//ContainsString  Helper functions to check from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

//RemoveString Helper function to check remove string
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func TrimSpaces(d interface{}) string {
	if s, ok := d.(string); ok {
		return strings.TrimSpace(s)
	}

	return ""
}

func TrimSpacesMap(m map[string]string) map[string]string {
	trimmed := map[string]string{}
	for key, v := range m {
		trimmed[key] = TrimSpaces(v)
	}
	return trimmed
}
