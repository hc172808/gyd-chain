import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { 
  CheckCircle, 
  XCircle, 
  Clock, 
  Server,
  Globe,
  Key,
  Trash2,
  RefreshCw
} from "lucide-react";
import { toast } from "sonner";

interface PendingNode {
  nodeId: string;
  hostname: string;
  publicIp: string;
  wireguardPubKey: string;
  registeredAt: string;
  type: string;
}

interface ApprovedNode {
  nodeId: string;
  hostname: string;
  publicIp: string;
  vpnAddress: string;
  approvedAt: string;
  lastSeen: string;
  syncHeight: number;
  status: "online" | "offline" | "syncing";
}

const mockPendingNodes: PendingNode[] = [
  {
    nodeId: "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
    hostname: "litenode-user1",
    publicIp: "203.0.113.45",
    wireguardPubKey: "WG_PUB_KEY_1...",
    registeredAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
    type: "litenode",
  },
  {
    nodeId: "b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef1234567a",
    hostname: "node-company-hq",
    publicIp: "198.51.100.22",
    wireguardPubKey: "WG_PUB_KEY_2...",
    registeredAt: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
    type: "litenode",
  },
  {
    nodeId: "c3d4e5f6789012345678901234567890abcdef1234567890abcdef1234567ab2",
    hostname: "validator-asia-1",
    publicIp: "192.0.2.100",
    wireguardPubKey: "WG_PUB_KEY_3...",
    registeredAt: new Date(Date.now() - 1000 * 60 * 60 * 5).toISOString(),
    type: "validator",
  },
];

const mockApprovedNodes: ApprovedNode[] = [
  {
    nodeId: "d4e5f6789012345678901234567890abcdef1234567890abcdef1234567ab2c3",
    hostname: "validator-eu-1",
    publicIp: "185.199.108.1",
    vpnAddress: "10.100.0.2",
    approvedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7).toISOString(),
    lastSeen: new Date(Date.now() - 1000 * 60 * 2).toISOString(),
    syncHeight: 1547820,
    status: "online",
  },
  {
    nodeId: "e5f6789012345678901234567890abcdef1234567890abcdef1234567ab2c3d4",
    hostname: "litenode-user2",
    publicIp: "203.0.113.100",
    vpnAddress: "10.100.0.3",
    approvedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString(),
    lastSeen: new Date(Date.now() - 1000 * 60 * 15).toISOString(),
    syncHeight: 1547800,
    status: "syncing",
  },
  {
    nodeId: "f6789012345678901234567890abcdef1234567890abcdef1234567ab2c3d4e5",
    hostname: "node-backup-1",
    publicIp: "198.51.100.50",
    vpnAddress: "10.100.0.4",
    approvedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 14).toISOString(),
    lastSeen: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
    syncHeight: 1547000,
    status: "offline",
  },
];

export function NodeManagement() {
  const [pendingNodes, setPendingNodes] = useState<PendingNode[]>(mockPendingNodes);
  const [approvedNodes, setApprovedNodes] = useState<ApprovedNode[]>(mockApprovedNodes);
  const [processingNode, setProcessingNode] = useState<string | null>(null);

  const handleApprove = async (nodeId: string) => {
    setProcessingNode(nodeId);
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      const node = pendingNodes.find(n => n.nodeId === nodeId);
      if (node) {
        setPendingNodes(prev => prev.filter(n => n.nodeId !== nodeId));
        setApprovedNodes(prev => [...prev, {
          nodeId: node.nodeId,
          hostname: node.hostname,
          publicIp: node.publicIp,
          vpnAddress: `10.100.0.${approvedNodes.length + 5}`,
          approvedAt: new Date().toISOString(),
          lastSeen: new Date().toISOString(),
          syncHeight: 0,
          status: "syncing",
        }]);
        toast.success(`Node ${node.hostname} approved and added to VPN`);
      }
    } catch {
      toast.error("Failed to approve node");
    } finally {
      setProcessingNode(null);
    }
  };

  const handleReject = async (nodeId: string) => {
    setProcessingNode(nodeId);
    
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      setPendingNodes(prev => prev.filter(n => n.nodeId !== nodeId));
      toast.success("Node rejected");
    } catch {
      toast.error("Failed to reject node");
    } finally {
      setProcessingNode(null);
    }
  };

  const handleRemove = async (nodeId: string) => {
    setProcessingNode(nodeId);
    
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      setApprovedNodes(prev => prev.filter(n => n.nodeId !== nodeId));
      toast.success("Node removed from network");
    } catch {
      toast.error("Failed to remove node");
    } finally {
      setProcessingNode(null);
    }
  };

  const formatTimeAgo = (dateString: string) => {
    const seconds = Math.floor((Date.now() - new Date(dateString).getTime()) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  };

  const getStatusBadge = (status: ApprovedNode["status"]) => {
    switch (status) {
      case "online":
        return <Badge className="bg-green-500/20 text-green-500 border-green-500/30">Online</Badge>;
      case "syncing":
        return <Badge className="bg-blue-500/20 text-blue-500 border-blue-500/30">Syncing</Badge>;
      case "offline":
        return <Badge className="bg-red-500/20 text-red-500 border-red-500/30">Offline</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      {/* Pending Nodes */}
      <Card className="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5 text-yellow-500" />
            Pending Node Requests
            {pendingNodes.length > 0 && (
              <Badge variant="secondary" className="ml-2">{pendingNodes.length}</Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {pendingNodes.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">No pending node requests</p>
          ) : (
            <div className="space-y-4">
              {pendingNodes.map((node) => (
                <div 
                  key={node.nodeId}
                  className="p-4 rounded-lg border border-border/50 bg-muted/20 space-y-3"
                >
                  <div className="flex items-start justify-between">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <Server className="h-4 w-4 text-primary" />
                        <span className="font-semibold">{node.hostname}</span>
                        <Badge variant="outline" className="text-xs">{node.type}</Badge>
                      </div>
                      <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Globe className="h-3 w-3" />
                          {node.publicIp}
                        </span>
                        <span className="flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {formatTimeAgo(node.registeredAt)}
                        </span>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        className="border-red-500/30 text-red-500 hover:bg-red-500/10"
                        onClick={() => handleReject(node.nodeId)}
                        disabled={processingNode === node.nodeId}
                      >
                        <XCircle className="h-4 w-4 mr-1" />
                        Reject
                      </Button>
                      <Button
                        size="sm"
                        className="bg-green-600 hover:bg-green-700"
                        onClick={() => handleApprove(node.nodeId)}
                        disabled={processingNode === node.nodeId}
                      >
                        {processingNode === node.nodeId ? (
                          <RefreshCw className="h-4 w-4 mr-1 animate-spin" />
                        ) : (
                          <CheckCircle className="h-4 w-4 mr-1" />
                        )}
                        Approve
                      </Button>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/30 p-2 rounded font-mono">
                    <Key className="h-3 w-3" />
                    Node ID: {node.nodeId.slice(0, 16)}...{node.nodeId.slice(-8)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Approved Nodes */}
      <Card className="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CheckCircle className="h-5 w-5 text-green-500" />
            Approved Nodes
            <Badge variant="secondary" className="ml-2">{approvedNodes.length}</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border/50">
                  <th className="text-left py-3 px-2 text-sm font-medium text-muted-foreground">Node</th>
                  <th className="text-left py-3 px-2 text-sm font-medium text-muted-foreground">VPN IP</th>
                  <th className="text-left py-3 px-2 text-sm font-medium text-muted-foreground">Public IP</th>
                  <th className="text-left py-3 px-2 text-sm font-medium text-muted-foreground">Status</th>
                  <th className="text-left py-3 px-2 text-sm font-medium text-muted-foreground">Sync Height</th>
                  <th className="text-left py-3 px-2 text-sm font-medium text-muted-foreground">Last Seen</th>
                  <th className="text-right py-3 px-2 text-sm font-medium text-muted-foreground">Actions</th>
                </tr>
              </thead>
              <tbody>
                {approvedNodes.map((node) => (
                  <tr key={node.nodeId} className="border-b border-border/30 hover:bg-muted/20">
                    <td className="py-3 px-2">
                      <div className="flex items-center gap-2">
                        <Server className="h-4 w-4 text-primary" />
                        <span className="font-medium">{node.hostname}</span>
                      </div>
                    </td>
                    <td className="py-3 px-2 font-mono text-sm">{node.vpnAddress}</td>
                    <td className="py-3 px-2 font-mono text-sm text-muted-foreground">{node.publicIp}</td>
                    <td className="py-3 px-2">{getStatusBadge(node.status)}</td>
                    <td className="py-3 px-2 font-mono text-sm">
                      {node.syncHeight.toLocaleString()}
                    </td>
                    <td className="py-3 px-2 text-sm text-muted-foreground">
                      {formatTimeAgo(node.lastSeen)}
                    </td>
                    <td className="py-3 px-2 text-right">
                      <Button
                        size="sm"
                        variant="ghost"
                        className="text-red-500 hover:text-red-600 hover:bg-red-500/10"
                        onClick={() => handleRemove(node.nodeId)}
                        disabled={processingNode === node.nodeId}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
