from diagrams import Cluster, Diagram

from diagrams.k8s.compute import Deployment, Pod, ReplicaSet
from diagrams.k8s.group import Namespace
from diagrams.k8s.network import Ingress, Service
from diagrams.k8s.others import CRD
from diagrams.k8s.rbac import ClusterRole, ClusterRoleBinding
from diagrams.saas.chat import Slack

with Diagram("Tugboat", show=False):
  slack = Slack()

  with Cluster():
    Namespace("tugboat")

    with Cluster("controller"):
      cpod = Pod("pod")

      ClusterRole("cr")
      ClusterRoleBinding("crb")
      CRD("releasehistory")
      Service("svc") >> cpod << ReplicaSet("rs") << Deployment("dep")

    with Cluster("notifier-slack"):
      nssvc = Service("svc")
      Ingress("ing") >> nssvc >> Pod("pod") << ReplicaSet("rs") << Deployment("dep")

    with Cluster("slack"):
      sing = Ingress("ing")
      sing >> Service("svc") >> Pod("pod") << ReplicaSet("rs") << Deployment("dep")

    with Cluster("watcher"):
      ClusterRole("cr")
      ClusterRoleBinding("crb")
      Service("svc") >> Pod("pod") << ReplicaSet("rs") << Deployment("dep")

    cpod >> nssvc

  slack >> sing 