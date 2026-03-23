package ovh

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	NotATaint TaintEffectType = iota
	NoExecute
	NoSchedule
	PreferNoSchedule
)

type CloudProjectKubeNodePoolCreateOpts struct {
	AntiAffinity      *bool                                `json:"antiAffinity,omitempty"`
	Autoscale         *bool                                `json:"autoscale,omitempty"`
	AvailabilityZones *[]string                            `json:"availabilityZones,omitempty"`
	DesiredNodes      *int                                 `json:"desiredNodes,omitempty"`
	FlavorName        string                               `json:"flavorName"`
	MaxNodes          *int                                 `json:"maxNodes,omitempty"`
	MinNodes          *int                                 `json:"minNodes,omitempty"`
	MonthlyBilled     *bool                                `json:"monthlyBilled,omitempty"`
	Name              *string                              `json:"name,omitempty"`
	Autoscaling       *CloudProjectKubeNodePoolAutoscaling `json:"autoscaling,omitempty"`
	Template          *CloudProjectKubeNodePoolTemplate    `json:"template,omitempty"`
}

type TaintEffectType int

type Taint struct {
	Effect TaintEffectType `json:"effect"`
	Key    string          `json:"key"`
	Value  string          `json:"value"`
}

type CloudProjectKubeNodePoolTemplateMetadata struct {
	Annotations map[string]string `json:"annotations"`
	Finalizers  []string          `json:"finalizers"`
	Labels      map[string]string `json:"labels"`
}

type CloudProjectKubeNodePoolTemplateSpec struct {
	Taints        []Taint `json:"taints"`
	Unschedulable bool    `json:"unschedulable"`
}

type CloudProjectKubeNodePoolTemplate struct {
	Metadata CloudProjectKubeNodePoolTemplateMetadata `json:"metadata"`
	Spec     CloudProjectKubeNodePoolTemplateSpec     `json:"spec"`
}

type CloudProjectKubeNodePoolAutoscaling struct {
	ScaleDownUtilizationThreshold *float64 `json:"scaleDownUtilizationThreshold,omitempty"`
	ScaleDownUnneededTimeSeconds  *int     `json:"scaleDownUnneededTimeSeconds,omitempty"`
	ScaleDownUnreadyTimeSeconds   *int     `json:"scaleDownUnreadyTimeSeconds,omitempty"`
}

type CloudProjectKubeNodePoolUpdateOpts struct {
	Autoscale    *bool                                `json:"autoscale,omitempty"`
	DesiredNodes *int                                 `json:"desiredNodes,omitempty"`
	MaxNodes     *int                                 `json:"maxNodes,omitempty"`
	MinNodes     *int                                 `json:"minNodes,omitempty"`
	Autoscaling  *CloudProjectKubeNodePoolAutoscaling `json:"autoscaling,omitempty"`
	Template     *CloudProjectKubeNodePoolTemplate    `json:"template,omitempty"`
}

var toString = map[TaintEffectType]string{
	NotATaint:        "",
	NoExecute:        "NoExecute",
	NoSchedule:       "NoSchedule",
	PreferNoSchedule: "PreferNoSchedule",
}

var TaintEffecTypeToID = map[string]TaintEffectType{
	"":                 NotATaint,
	"NoExecute":        NoExecute,
	"NoSchedule":       NoSchedule,
	"PreferNoSchedule": PreferNoSchedule,
}

func (e TaintEffectType) String() string {
	return toString[e]
}

// MarshalJSON marshals the enum as a quoted json string
func (e TaintEffectType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[e])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (e *TaintEffectType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*e = TaintEffecTypeToID[j]
	return nil
}

type CloudProjectKubeNodePoolResponse struct {
	Autoscale         bool                                `json:"autoscale"`
	AntiAffinity      bool                                `json:"antiAffinity"`
	AvailabilityZones []string                            `json:"availabilityZones"`
	AvailableNodes    int                                 `json:"availableNodes"`
	CreatedAt         string                              `json:"createdAt"`
	CurrentNodes      int                                 `json:"currentNodes"`
	DesiredNodes      int                                 `json:"desiredNodes"`
	Flavor            string                              `json:"flavor"`
	Id                string                              `json:"id"`
	MaxNodes          int                                 `json:"maxNodes"`
	MinNodes          int                                 `json:"minNodes"`
	MonthlyBilled     bool                                `json:"monthlyBilled"`
	Name              string                              `json:"name"`
	ProjectId         string                              `json:"projectId"`
	SizeStatus        string                              `json:"sizeStatus"`
	Status            string                              `json:"status"`
	UpToDateNodes     int                                 `json:"upToDateNodes"`
	UpdatedAt         string                              `json:"updatedAt"`
	Autoscaling       CloudProjectKubeNodePoolAutoscaling `json:"autoscaling"`
	Template          *CloudProjectKubeNodePoolTemplate   `json:"template,omitempty"`
}

func (n *CloudProjectKubeNodePoolResponse) String() string {
	return fmt.Sprintf("%s(%s): %s", n.Name, n.Id, n.Status)
}
