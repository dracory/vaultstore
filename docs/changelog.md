# Changelog

This document contains the version history and changes for the VaultStore project.

## 2026

### 2026.02.03
- Added identity-based password management for optimized bulk rekey operations
- Added configurable cryptographic parameters via CryptoConfig (Argon2id, AES-256-GCM)
- Added vault_meta table for password identity tracking and vault settings
- Added BulkRekey and MigrateRecordLinks methods for efficient password rotation
- Updated encryption from custom XOR scheme to AES-256-GCM with Argon2id KDF

## 2025

### 2025.03.13
- Added soft delete functionality and token soft delete method

### 2025.03.12
- Added record query interface and record interface
- Improved test coverage

### 2025.01.15
- Updated store interface

## 2024

### 2024.12.07
- Added context to interfaced dependencies

### 2024.11.28
- Upgraded dependencies

### 2024.11.01
- Improved efficiency and speed
- Added benchmarks

### 2024.09.15
- Updated to use math/rand/v2 package

### 2024.09.11
- Added tokens read functionality

### 2024.08.20
- Added token exists and record count methods

### 2024.08.11
- Added token support, to allow tokenization
- More info at: https://en.wikipedia.org/wiki/Tokenization_(data_security)

## 2022

### 2022.12.26
- Updated ID to Human Friendly UUID

## 2021

### 2021.12.28
- Removed GORM dependency and moved to the standard library
- Fixed bug where DB scanner was returning empty values

### 2021.12.12
- Added tests badge
