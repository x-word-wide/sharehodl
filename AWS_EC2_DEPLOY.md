# ShareHODL AWS EC2 Deployment Guide

## 🚀 Deploy ShareHODL on AWS EC2

Perfect choice! AWS EC2 provides enterprise-grade infrastructure with excellent performance and reliability.

## Instance Requirements

### Recommended: `t3.xlarge`
- **vCPUs**: 4 (Intel Xeon Platinum 8000)
- **Memory**: 16 GiB
- **Network**: Up to 5 Gbps
- **EBS**: 500 GB gp3 (3,000 IOPS)
- **Cost**: ~$120/month (On-Demand)
- **Cost**: ~$70/month (1-year Reserved)

### Budget Option: `t3.large` 
- **vCPUs**: 2
- **Memory**: 8 GiB
- **Network**: Up to 5 Gbps
- **EBS**: 200 GB gp3
- **Cost**: ~$60/month (On-Demand)
- **Cost**: ~$35/month (1-year Reserved)

## Quick Launch Instructions

### Step 1: Launch EC2 Instance

**Via AWS Console:**
1. Go to EC2 Dashboard
2. Click "Launch Instance"
3. **AMI**: Ubuntu Server 22.04 LTS
4. **Instance Type**: `t3.xlarge` (recommended)
5. **Storage**: 500 GB gp3 SSD
6. **Security Group**: Allow ports 22, 80, 443, 26657, 1317
7. **Key Pair**: Create or select existing
8. Launch instance

**Via AWS CLI:**
```bash
# Create security group
aws ec2 create-security-group \
  --group-name sharehodl-sg \
  --description "ShareHODL Testnet Security Group"

# Add firewall rules
aws ec2 authorize-security-group-ingress \
  --group-name sharehodl-sg \
  --protocol tcp --port 22 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-name sharehodl-sg \
  --protocol tcp --port 80 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-name sharehodl-sg \
  --protocol tcp --port 443 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-name sharehodl-sg \
  --protocol tcp --port 26657 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-name sharehodl-sg \
  --protocol tcp --port 1317 --cidr 0.0.0.0/0

# Launch instance
aws ec2 run-instances \
  --image-id ami-0c7217cdde317cfec \
  --count 1 \
  --instance-type t3.xlarge \
  --key-name YOUR_KEY_NAME \
  --security-groups sharehodl-sg \
  --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"VolumeSize":500,"VolumeType":"gp3"}}]' \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=ShareHODL-Testnet}]'
```

### Step 2: Deploy ShareHODL

**SSH to your instance:**
```bash
# Get instance public IP from AWS console
ssh -i your-key.pem ubuntu@YOUR_EC2_PUBLIC_IP

# Download and run deployment script
wget https://raw.githubusercontent.com/sharehodl/sharehodl-blockchain/main/scripts/deploy-aws-ec2.sh
chmod +x deploy-aws-ec2.sh
sudo ./deploy-aws-ec2.sh
```

**The script automatically:**
- ✅ Installs all dependencies
- ✅ Configures AWS CloudWatch monitoring
- ✅ Builds ShareHODL blockchain
- ✅ Sets up single-node testnet
- ✅ Deploys all 5 frontend applications
- ✅ Configures Nginx reverse proxy
- ✅ Sets up security groups
- ✅ Creates systemd services
- ✅ Enables SSL-ready configuration

## AWS-Specific Features

### CloudWatch Integration
**Automatic monitoring of:**
- CPU, Memory, Disk usage
- Application logs
- Blockchain metrics
- Frontend performance

**View in AWS Console:**
- CloudWatch → Logs → `sharehodl-{instance-id}`
- CloudWatch → Metrics → `ShareHODL/EC2`

### IAM Permissions (Optional)
**For enhanced AWS integration:**
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:CreateTags",
                "ec2:AuthorizeSecurityGroupIngress",
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
```

### Elastic IP (Recommended)
**Allocate static IP:**
```bash
# Allocate Elastic IP
aws ec2 allocate-address --domain vpc

# Associate with instance
aws ec2 associate-address \
  --instance-id i-1234567890abcdef0 \
  --allocation-id eipalloc-12345678
```

## Access Your ShareHODL Testnet

**After deployment (15-20 minutes):**
- **Main Portal**: `http://YOUR_EC2_IP/`
- **Governance**: `http://YOUR_EC2_IP/governance`
- **Trading**: `http://YOUR_EC2_IP/trading`
- **Explorer**: `http://YOUR_EC2_IP/explorer`
- **Business**: `http://YOUR_EC2_IP/business`
- **API**: `http://YOUR_EC2_IP/api/node_info`
- **Health Check**: `http://YOUR_EC2_IP/health`

## Cost Optimization

### Reserved Instances
**Save 60% with 1-year commitment:**
- `t3.xlarge`: $70/month (vs $120 On-Demand)
- `t3.large`: $35/month (vs $60 On-Demand)

### Spot Instances
**Save up to 90% for development:**
- `t3.xlarge Spot`: ~$25-40/month
- **Note**: May be interrupted

### Auto Scaling (Future)
**Scale based on usage:**
```bash
# Create Auto Scaling Group for multiple validators
aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name sharehodl-validators \
  --min-size 1 \
  --max-size 5 \
  --desired-capacity 3
```

## Production Enhancements

### Application Load Balancer
**For high availability:**
- Route traffic to multiple validators
- Health checks on `/health` endpoint
- SSL termination
- Cross-AZ deployment

### RDS for Metadata
**External database for analytics:**
- Aurora PostgreSQL
- Store trading history
- User preferences
- Analytics data

### ElastiCache
**Redis for caching:**
- Session storage
- API response caching
- Real-time data

## Monitoring & Alerts

### CloudWatch Alarms
**Automatic alerts for:**
```bash
# High CPU usage
aws cloudwatch put-metric-alarm \
  --alarm-name "ShareHODL-HighCPU" \
  --alarm-description "ShareHODL high CPU usage" \
  --metric-name CPUUtilization \
  --namespace AWS/EC2 \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold

# Blockchain node down
aws cloudwatch put-metric-alarm \
  --alarm-name "ShareHODL-NodeDown" \
  --alarm-description "ShareHODL blockchain node down" \
  --metric-name StatusCheckFailed \
  --namespace AWS/EC2 \
  --statistic Maximum \
  --period 60 \
  --threshold 1 \
  --comparison-operator GreaterThanOrEqualToThreshold
```

## SSL Certificate (Production)

### Route 53 + Certificate Manager
**Free SSL certificates:**
```bash
# Request certificate
aws acm request-certificate \
  --domain-name sharehodl.yourdomain.com \
  --validation-method DNS

# Update nginx configuration for HTTPS
sudo certbot --nginx -d sharehodl.yourdomain.com
```

## Backup Strategy

### EBS Snapshots
**Automated backups:**
```bash
# Create snapshot
aws ec2 create-snapshot \
  --volume-id vol-1234567890abcdef0 \
  --description "ShareHODL daily backup"

# Schedule with Lambda function for daily backups
```

### S3 Backup
**Blockchain data backup:**
```bash
# Backup blockchain data to S3
aws s3 sync ~/.sharehodl/data/ s3://sharehodl-backup/$(date +%Y-%m-%d)/
```

## Total AWS Costs

### Monthly Costs (t3.xlarge Reserved):
- **EC2 Instance**: $70/month
- **EBS Storage**: $40/month (500GB gp3)
- **Data Transfer**: $5-15/month
- **CloudWatch**: $5/month
- **Elastic IP**: $3.65/month
- **Total**: ~$125/month

**Much more reliable than basic VPS with enterprise features!**

## Advanced AWS Features

### Auto Scaling (Multi-Validator)
- Scale validators based on load
- Cross-AZ deployment
- Auto-recovery

### Global Deployment  
- Multiple regions
- CloudFront CDN
- Route 53 DNS

### Security
- WAF protection
- VPC isolation  
- IAM roles
- Secrets Manager

**AWS EC2 provides the perfect foundation for ShareHODL's professional infrastructure!** 🚀