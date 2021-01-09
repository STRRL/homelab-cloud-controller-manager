package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	iputil "github.com/strrl/homelab-cloud-controller-manager/pkg/ip"
	corev1 "k8s.io/api/core/v1"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const SubNetAnnotation = "hccm.strrl.dev/subnet"

type ServiceLoadBalancerReconciler struct {
	logger        logr.Logger
	kubeClient    client.Client
	clusterIPMask net.IPMask
}

func NewServiceLoadBalancerReconciler(client client.Client, logger logr.Logger) *ServiceLoadBalancerReconciler {
	return &ServiceLoadBalancerReconciler{
		kubeClient: client,
		logger:     logger,
		// TODO: hard-coded default 10.96.0.0/12
		clusterIPMask: net.IPv4Mask(255, 240, 0, 0),
	}
}

func (it *ServiceLoadBalancerReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	it.logger.WithValues("request", request).Info("reconcile request received")
	service := corev1.Service{}
	err := it.kubeClient.Get(context.Background(), request.NamespacedName, &service)

	if err != nil {
		it.logger.Error(err, "failed to fetch target service")
		// err is nil, will not requeue
		return reconcile.Result{}, nil
	}

	if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		it.logger.WithValues("service", request.NamespacedName, "service type", service.Spec.Type).Info("service type is not LoadBalancer, skipping")
		return reconcile.Result{}, nil
	}

	clusterIp := service.Spec.ClusterIP
	if len(clusterIp) == 0 {
		return reconcile.Result{}, nil
	}

	if _, ok := service.GetAnnotations()[SubNetAnnotation]; !ok {
		it.logger.Info(fmt.Sprintf("service do not contains annotation %s, skipping", SubNetAnnotation), "key", request.NamespacedName)
		return reconcile.Result{}, nil
	}

	externalIpNet := service.GetAnnotations()[SubNetAnnotation]
	externalIp, err := it.generateExternalIp(clusterIp, externalIpNet)

	if err != nil {
		// wrong with configurations, will not requeue
		return reconcile.Result{}, nil
	}

	alreadyContains := false
	updated := service.DeepCopy()
	for _, item := range updated.Status.LoadBalancer.Ingress {
		if item.IP == externalIp {
			it.logger.Info("ingress ip already exist", "service", request.NamespacedName, "ip", externalIp)
			alreadyContains = true
		}
	}

	if !alreadyContains {
		it.logger.Info("update ingress", "service", request.NamespacedName, "ingress-ip", externalIp)
		// it will replace all your ingress IP
		updated.Status.LoadBalancer.Ingress = append([]corev1.LoadBalancerIngress(nil), corev1.LoadBalancerIngress{
			IP: externalIp,
		})

		err := it.kubeClient.Status().Update(context.Background(), updated)
		if err != nil {
			it.logger.Error(err, "failed to update service")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (it *ServiceLoadBalancerReconciler) generateExternalIp(clusterIp, externalIpNet string) (externalIp string, err error) {
	_, ipNet, err := net.ParseCIDR(externalIpNet)
	if err != nil {
		it.logger.WithValues(SubNetAnnotation, externalIpNet).Error(err, "failed to parse CIDR from %s", SubNetAnnotation)
		// err is nil, will not requeue
		return "", nil
	}
	maskSize, _ := it.clusterIPMask.Size()
	_, hostBits, err := iputil.SplitWithMask(clusterIp, uint(maskSize))
	if err != nil {
		it.logger.Error(err, "split ClusterIP failed")

		return "", err
	}
	newNetBit, err := iputil.Ip2long(ipNet.IP.String())
	if err != nil {
		it.logger.Error(err, "parse new netBit failed")
		// err is nil, will not requeue
		return "", err
	}
	return iputil.Long2ip(hostBits | newNetBit), nil
}
