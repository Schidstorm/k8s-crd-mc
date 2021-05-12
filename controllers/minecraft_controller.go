/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	kapps "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	minecraftv1 "k8s-operators/api/v1"
)

var (
	podOwnerKey           = ".metadata.controller"
	apiGVStr              = minecraftv1.GroupVersion.String()
	requeueResult         = ctrl.Result{RequeueAfter: 2 * time.Second}
	minecraftPodImageName = "itzg/minecraft-server:latest"
)

// MinecraftReconciler reconciles a Minecraft object
type MinecraftReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=minecraft.schidlow.ski,resources=minecrafts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=minecraft.schidlow.ski,resources=minecrafts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=minecraft.schidlow.ski,resources=minecrafts/finalizers,verbs=update
//+kubebuilder:rbac:groups=minecraft.schidlow.ski,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=minecraft.schidlow.ski,resources=pods/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Minecraft object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MinecraftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("minecraft", req.NamespacedName)

	var minecraft minecraftv1.Minecraft
	if err := r.Get(ctx, req.NamespacedName, &minecraft); err != nil {
		return requeueResult, client.IgnoreNotFound(err)
	}

	var childPods kapps.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingFields{podOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child Jobs")
		return requeueResult, err
	}

	podAvailable := false
	minecraft.Status.Pod = "a"
	for _, pod := range childPods.Items {
		if hasPodStopped(pod) {
			continue
		}

		podAvailable = true
		minecraft.Status.Pod = pod.Name
		minecraft.Status.Status = string(pod.Status.Phase)
		break
	}

	if !podAvailable {
		minecraft.Status.Status = "None"
	}

	if err := r.Status().Update(ctx, &minecraft); err != nil {
		log.Error(err, "unable to update status")
		return requeueResult, err
	}

	for _, pod := range childPods.Items {
		if podAvailable {
			if minecraft.Status.Pod == pod.Name {
				continue
			}
		}
		if err := r.Client.Delete(ctx, &pod); err != nil {
			log.Error(err, "Pod deletion failed")
			return requeueResult, err
		}
	}

	if !podAvailable {
		pod := GeneratePodFromSpec(minecraft.Spec)
		pod.ObjectMeta.Namespace = minecraft.Namespace
		pod.ObjectMeta.GenerateName = "minecraft-pod-"
		if err := ctrl.SetControllerReference(&minecraft, pod, r.Scheme); err != nil {
			log.Error(err, "unable to set controller reference")
			return requeueResult, err
		}

		if err := r.Create(ctx, pod); err != nil {
			log.Error(err, "unable to create pod")
			return requeueResult, err
		}

		if err := r.Status().Update(ctx, &minecraft); err != nil {
			log.Error(err, "unable to update status")
			return requeueResult, err
		}
	}

	return requeueResult, nil
}

func GeneratePodFromSpec(spec minecraftv1.MinecraftSpec) *kapps.Pod {
	envs := []kapps.EnvVar{
		{Name: "EULA", Value: "true"},
		{Name: "ALLOW_NETHER", Value: "true"},
		{Name: "ANNOUNCE_PLAYER_ACHIEVEMENTS", Value: "true"},
		{Name: "GENERATE_STRUCTURES", Value: "true"},
		{Name: "SPAWN_ANIMALS", Value: "true"},
		{Name: "ALLOW_FLIGHT", Value: "true"},
	}

	containerPort := uint16(25565)
	if spec.Ports != nil && spec.Ports.Minecraft != nil {
		envs = append(envs, kapps.EnvVar{Name: "SERVER_PORT", Value: strconv.Itoa(int(*spec.Ports.Minecraft))})
		containerPort = *spec.Ports.Minecraft
	}
	if spec.Mode != nil {
		envs = append(envs, kapps.EnvVar{Name: "MODE", Value: string(*spec.Mode)})
	}
	if spec.Name != nil {
		envs = append(envs, kapps.EnvVar{Name: "SERVER_NAME", Value: *spec.Name})
	}
	if spec.Difficulty != nil {
		envs = append(envs, kapps.EnvVar{Name: "DIFFICULTY", Value: string(*spec.Difficulty)})
	}
	if spec.Motd != nil {
		envs = append(envs, kapps.EnvVar{Name: "MOTD", Value: *spec.Motd})
	}
	if spec.Seed != nil {
		envs = append(envs, kapps.EnvVar{Name: "SEED", Value: *spec.Seed})
	}

	return &kapps.Pod{
		ObjectMeta: spec.Template,
		Spec: kapps.PodSpec{
			Containers: []kapps.Container{
				{
					Name:  "minecraft-server",
					Image: minecraftPodImageName,
					Resources: kapps.ResourceRequirements{
						Limits: map[kapps.ResourceName]resource.Quantity{
							"memory": resource.MustParse("2Gi"),
						},
						Requests: map[kapps.ResourceName]resource.Quantity{
							"memory": resource.MustParse("1.5Gi"),
						},
					},
					Env: envs,
					Ports: []kapps.ContainerPort{
						{ContainerPort: int32(containerPort), Name: "minecraft"},
					},
					ReadinessProbe: &kapps.Probe{
						Handler: kapps.Handler{
							Exec: &kapps.ExecAction{Command: []string{"mcstatus", "127.0.0.1", "ping"}},
						},
						InitialDelaySeconds: 30,
						PeriodSeconds:       30,
					},
					LivenessProbe: &kapps.Probe{
						Handler: kapps.Handler{
							Exec: &kapps.ExecAction{Command: []string{"mcstatus", "127.0.0.1", "ping"}},
						},
						InitialDelaySeconds: 30,
						PeriodSeconds:       30,
					},
				},
			},
		},
	}
}

func hasPodStopped(pod kapps.Pod) bool {
	return pod.Status.Phase != kapps.PodPending && pod.Status.Phase != kapps.PodRunning
}

// SetupWithManager sets up the controller with the Manager.
func (r *MinecraftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kapps.Pod{}, podOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*kapps.Pod)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "Minecraft" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&minecraftv1.Minecraft{}).
		Complete(r)
}
