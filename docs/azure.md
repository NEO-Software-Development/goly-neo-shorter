# Deploying Goly to Azure Kubernetes Service (AKS)

This guide provides high-level instructions for deploying the Goly application to Azure Kubernetes Service (AKS) with a managed PostgreSQL database.

## Introduction

This guide assumes you have a basic understanding of Azure, Kubernetes, and Terraform. The deployment process involves:

1.  Setting up an Azure account and the necessary tools.
2.  Creating an AKS cluster and a PostgreSQL database on Azure.
3.  Using Terraform to deploy the Goly application to the AKS cluster.

## Prerequisites

You will need the following tools installed on your local machine:

-   **Azure CLI:** The command-line interface for Azure. You can find installation instructions [here](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli).
-   **kubectl:** The Kubernetes command-line tool. You can install it using the Azure CLI:
    ```bash
    az aks install-cli
    ```
-   **Terraform:** The infrastructure as code tool. You can find installation instructions [here](https://learn.hashicorp.com/tutorials/terraform/install-cli).

## Azure Setup

1.  **Create an Azure account:** If you don't have one already, you can create a free Azure account [here](https://azure.microsoft.com/en-us/free/).
2.  **Log in to Azure:**
    ```bash
    az login
    ```

## Create Azure Resources

You will need to create a resource group, an AKS cluster, and a PostgreSQL database.

1.  **Create a resource group:**
    ```bash
    az group create --name goly-resources --location eastus
    ```

2.  **Create an AKS cluster:** We'll use a small and cheap VM size to keep costs down.
    ```bash
    az aks create \
        --resource-group goly-resources \
        --name goly-cluster \
        --node-count 1 \
        --node-vm-size Standard_B2s \
        --generate-ssh-keys
    ```

3.  **Get AKS credentials:**
    ```bash
    az aks get-credentials --resource-group goly-resources --name goly-cluster
    ```

4.  **Create an Azure Database for PostgreSQL:**
    ```bash
    az postgres server create \
        --resource-group goly-resources \
        --name goly-postgres-server \
        --location eastus \
        --admin-user <your-admin-user> \
        --admin-password <your-admin-password> \
        --sku-name B_Gen5_1 \
        --version 11
    ```
    **Note:** Replace `<your-admin-user>` and `<your-admin-password>` with your own credentials.

5.  **Configure PostgreSQL firewall:** You will need to configure the firewall to allow connections from your AKS cluster. This is a complex topic that depends on your network configuration. For a quick start, you can allow all Azure services to connect:
    ```bash
    az postgres server firewall-rule create \
        --resource-group goly-resources \
        --server goly-postgres-server \
        --name AllowAzureServices \
        --start-ip-address 0.0.0.0 \
        --end-ip-address 0.0.0.0
    ```
    **Warning:** This is not a secure configuration for a production environment. You should configure more restrictive firewall rules.

## Terraform Deployment

Once the Azure resources are created, you can use Terraform to deploy the Goly application.

1.  **Navigate to the Terraform directory:**
    ```bash
    cd infra/terraform
    ```

2.  **Create a `terraform.tfvars` file:** This file will contain the connection details for your PostgreSQL database.
    ```
    db_host = "goly-postgres-server.postgres.database.azure.com"
    db_user = "<your-admin-user>"
    db_password = "<your-admin-password>"
    db_name = "postgres" # Or the name of the database you created
    ```

3.  **Initialize Terraform:**
    ```bash
    terraform init
    ```

4.  **Apply the Terraform configuration:**
    ```bash
    terraform apply
    ```

This will deploy the Goly application to your AKS cluster.

## Post-Deployment Steps

1.  **Get the public IP address:**
    ```bash
    kubectl get service goly-service -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
    ```

2.  **Access the application:** You can now access the application by navigating to the public IP address in your web browser.

This guide provides a high-level overview of the deployment process. You may need to adjust the steps based on your specific requirements and environment.
