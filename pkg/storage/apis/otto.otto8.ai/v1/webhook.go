package v1

import (
	"github.com/otto8-ai/otto8/apiclient/types"
	"github.com/otto8-ai/otto8/pkg/system"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ Aliasable = (*Webhook)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookSpec   `json:"spec,omitempty"`
	Status WebhookStatus `json:"status,omitempty"`
}

func (w *Webhook) GetAliasName() string {
	return w.Spec.WebhookManifest.Alias
}

func (w *Webhook) SetAssigned() {
	w.Status.AliasAssigned = true
}

func (w *Webhook) IsAssigned() bool {
	return w.Status.AliasAssigned
}

func (*Webhook) GetColumns() [][]string {
	return [][]string{
		{"Name", "Name"},
		{"Alias", "Spec.Alias"},
		{"Workflow", "Spec.Workflow"},
		{"Created", "{{ago .CreationTimestamp}}"},
		{"Last Success", "{{ago .Status.LastSuccessfulRunCompleted}}"},
		{"Description", "Spec.Description"},
	}
}

func (w *Webhook) DeleteRefs() []Ref {
	if system.IsWebhookID(w.Spec.Workflow) {
		return []Ref{
			{ObjType: new(Workflow), Name: w.Spec.Workflow},
		}
	}
	return nil
}

type WebhookSpec struct {
	types.WebhookManifest `json:",inline"`
	TokenHash             []byte `json:"tokenHash,omitempty"`
}

type WebhookStatus struct {
	Alias                      string       `json:"alias,omitempty"`
	AliasAssigned              bool         `json:"aliasAssigned,omitempty"`
	LastSuccessfulRunCompleted *metav1.Time `json:"lastSuccessfulRunCompleted,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webhook `json:"items"`
}
