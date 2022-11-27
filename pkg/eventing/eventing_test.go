package eventing

import (
	"testing"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestHasExistingEventingInstallation(t *testing.T) {
	logger := logr.Discard()
	tests := []struct {
		name    string
		c       client.Client
		want    bool
		wantErr bool
	}{
		{
			name:    "No deployments on the cluster",
			c:       fake.NewFakeClient(),
			want:    false,
			wantErr: false,
		},
		{
			name: "No deployment on the cluster matching the Eventing Labels",
			c: fake.NewFakeClient(
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "d1", Labels: map[string]string{"test": "test"}},
				},
			),
			want:    false,
			wantErr: false,
		},
		{
			name: "One deployment with Eventing matching Labels",
			c: fake.NewFakeClient(
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "d1", Labels: eventingLabels},
				},
			),
			want:    true,
			wantErr: false,
		},
		{
			name: "Multiple deployments with Eventing matching Labels",
			c: fake.NewFakeClient(
				&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "d1", Labels: eventingLabels}},
				&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "d2", Labels: eventingLabels}},
			),
			want:    true,
			wantErr: false,
		},
		{
			name: "One deployment with partially matching labels",
			c: fake.NewFakeClient(
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "d1", Labels: map[string]string{"app": "eventing-operator", "test": "test"}},
				},
			),
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isInstalledWithClient(tt.c, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasExistingEventingInstallation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HasExistingEventingInstallation() = %v, want %v", got, tt.want)
			}
		})
	}
}
