import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { 
  GitBranch, 
  RefreshCw, 
  Server, 
  Shield, 
  CheckCircle, 
  XCircle, 
  Clock,
  Activity,
  HardDrive,
  Cpu,
  Users
} from "lucide-react";
import { NodeManagement } from "./NodeManagement";
import { toast } from "sonner";

interface SystemStatus {
  nodeStatus: "running" | "stopped" | "error";
  indexerStatus: "running" | "stopped" | "error";
  pendingNodes: number;
  approvedNodes: number;
  blockHeight: number;
  peerCount: number;
  uptime: string;
}

const mockSystemStatus: SystemStatus = {
  nodeStatus: "running",
  indexerStatus: "running",
  pendingNodes: 3,
  approvedNodes: 12,
  blockHeight: 1547823,
  peerCount: 24,
  uptime: "14d 7h 32m",
};

export function AdminDashboard() {
  const [isUpdating, setIsUpdating] = useState(false);
  const [isRebuilding, setIsRebuilding] = useState(false);
  const [systemStatus] = useState<SystemStatus>(mockSystemStatus);

  const handleGitPull = async () => {
    setIsUpdating(true);
    toast.info("Pulling latest changes from GitHub...");
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 3000));
      toast.success("Successfully pulled and rebuilt from GitHub!");
    } catch {
      toast.error("Failed to update from GitHub");
    } finally {
      setIsUpdating(false);
    }
  };

  const handleRebuildFrontend = async () => {
    setIsRebuilding(true);
    toast.info("Rebuilding frontend...");
    
    try {
      await new Promise(resolve => setTimeout(resolve, 2000));
      toast.success("Frontend rebuilt successfully!");
    } catch {
      toast.error("Failed to rebuild frontend");
    } finally {
      setIsRebuilding(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-500";
      case "stopped":
        return "bg-yellow-500";
      case "error":
        return "bg-red-500";
      default:
        return "bg-gray-500";
    }
  };

  return (
    <div className="container mx-auto max-w-6xl px-4 py-6">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Admin Dashboard</h1>
            <p className="text-muted-foreground">Manage GYDS Chain infrastructure</p>
          </div>
          <div className="flex gap-3">
            <Button
              onClick={handleRebuildFrontend}
              disabled={isRebuilding}
              variant="outline"
              className="gap-2"
            >
              <RefreshCw className={`h-4 w-4 ${isRebuilding ? "animate-spin" : ""}`} />
              Rebuild Frontend
            </Button>
            <Button
              onClick={handleGitPull}
              disabled={isUpdating}
              className="gap-2 bg-primary"
            >
              <GitBranch className={`h-4 w-4 ${isUpdating ? "animate-pulse" : ""}`} />
              {isUpdating ? "Updating..." : "Pull & Rebuild"}
            </Button>
          </div>
        </div>

        {/* Status Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card className="bg-card/50 backdrop-blur-sm border-border/50">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Node Status
              </CardTitle>
              <Server className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2">
                <div className={`h-2 w-2 rounded-full ${getStatusColor(systemStatus.nodeStatus)}`} />
                <span className="text-2xl font-bold capitalize">{systemStatus.nodeStatus}</span>
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Block #{systemStatus.blockHeight.toLocaleString()}
              </p>
            </CardContent>
          </Card>

          <Card className="bg-card/50 backdrop-blur-sm border-border/50">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Pending Nodes
              </CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-yellow-500">
                {systemStatus.pendingNodes}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Awaiting approval
              </p>
            </CardContent>
          </Card>

          <Card className="bg-card/50 backdrop-blur-sm border-border/50">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Active Nodes
              </CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-500">
                {systemStatus.approvedNodes}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Connected to VPN
              </p>
            </CardContent>
          </Card>

          <Card className="bg-card/50 backdrop-blur-sm border-border/50">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Uptime
              </CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{systemStatus.uptime}</div>
              <p className="text-xs text-muted-foreground mt-1">
                {systemStatus.peerCount} peers connected
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Main Content Tabs */}
        <Tabs defaultValue="nodes" className="space-y-4">
          <TabsList className="bg-muted/50">
            <TabsTrigger value="nodes" className="gap-2">
              <Server className="h-4 w-4" />
              Node Management
            </TabsTrigger>
            <TabsTrigger value="system" className="gap-2">
              <Cpu className="h-4 w-4" />
              System
            </TabsTrigger>
            <TabsTrigger value="security" className="gap-2">
              <Shield className="h-4 w-4" />
              Security
            </TabsTrigger>
          </TabsList>

          <TabsContent value="nodes">
            <NodeManagement />
          </TabsContent>

          <TabsContent value="system">
            <SystemPanel />
          </TabsContent>

          <TabsContent value="security">
            <SecurityPanel />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

function SystemPanel() {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Card className="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <HardDrive className="h-5 w-5" />
            Services
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {[
            { name: "gydschain-node", status: "active" },
            { name: "gydschain-indexer", status: "active" },
            { name: "gydschain-admin", status: "active" },
            { name: "nginx", status: "active" },
            { name: "postgresql", status: "active" },
            { name: "wg-quick@wg0", status: "active" },
          ].map((service) => (
            <div key={service.name} className="flex items-center justify-between p-2 rounded-lg bg-muted/30">
              <span className="font-mono text-sm">{service.name}</span>
              <Badge variant={service.status === "active" ? "default" : "secondary"} className="gap-1">
                {service.status === "active" ? (
                  <CheckCircle className="h-3 w-3" />
                ) : (
                  <XCircle className="h-3 w-3" />
                )}
                {service.status}
              </Badge>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card className="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Resource Usage
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <div className="flex justify-between text-sm mb-1">
              <span>CPU</span>
              <span>23%</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div className="h-full w-[23%] bg-primary rounded-full" />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-sm mb-1">
              <span>Memory</span>
              <span>4.2 GB / 16 GB</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div className="h-full w-[26%] bg-primary rounded-full" />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-sm mb-1">
              <span>Disk</span>
              <span>128 GB / 500 GB</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div className="h-full w-[25%] bg-primary rounded-full" />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-sm mb-1">
              <span>Network</span>
              <span>12.4 MB/s</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div className="h-full w-[45%] bg-green-500 rounded-full" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function SecurityPanel() {
  return (
    <Card className="bg-card/50 backdrop-blur-sm border-border/50">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          Security Overview
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
          <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
            <p className="text-green-500 font-medium">Firewall</p>
            <p className="text-2xl font-bold text-foreground">Active</p>
            <p className="text-sm text-muted-foreground">UFW enabled</p>
          </div>
          <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
            <p className="text-green-500 font-medium">VPN</p>
            <p className="text-2xl font-bold text-foreground">Secured</p>
            <p className="text-sm text-muted-foreground">WireGuard active</p>
          </div>
          <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
            <p className="text-green-500 font-medium">Fail2Ban</p>
            <p className="text-2xl font-bold text-foreground">Protected</p>
            <p className="text-sm text-muted-foreground">3 jails active</p>
          </div>
        </div>

        <div className="space-y-2">
          <h4 className="font-medium">Recent Security Events</h4>
          <div className="space-y-2">
            {[
              { event: "SSH login from 192.168.1.100", time: "2 min ago", type: "info" },
              { event: "Blocked IP: 45.33.22.11", time: "15 min ago", type: "warning" },
              { event: "Node approved: lite-node-7", time: "1 hour ago", type: "success" },
            ].map((event, i) => (
              <div key={i} className="flex items-center justify-between p-3 rounded-lg bg-muted/30">
                <span className="text-sm">{event.event}</span>
                <span className="text-xs text-muted-foreground">{event.time}</span>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
