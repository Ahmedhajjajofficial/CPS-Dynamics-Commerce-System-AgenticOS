# مواصفة توحيد أنماط الكود

## Why
يوجد في المشروع عدم اتساق كبير في أنماط الكود بين الأقسام المختلفة (React/TypeScript، Python، Go) وداخل نفس القسم. هذا يؤدي إلى:
- صعوبة في قراءة وفهم الكود
- صعوبة في الصيانة والتطوير
- تكلفة أعلى للإعداد

## What Changes
- توحيد تنسيق التعليقات الرأسية في جميع الملفات
- توحيد اتفاقية تسمية الملفات في React (PascalCase للمكونات)
- إضافة/توحيد استخدام `from __future__ import annotations` في Python
- إنشاء Style Guide موحد للمشروع
- إعداد أدوات linting متسقة

**BREAKING**: قد تتغير مسارات بعض الاستيرادات إذا تم تغيير أسماء الملفات

## Impact
- Affected specs: جميع مواصفات المشروع
- Affected code: 
  - `/workspace/app/src/components/ui/*.tsx`
  - `/workspace/cps-enterprise-dcs/local-agent/src/*.py`
  - `/workspace/cps-enterprise-dcs/regional-agent/internal/**/*.go`

## ADDED Requirements

### Requirement: توحيد التعليقات الرأسية
The system SHALL يوفر تعليقات رأسية متسقة في جميع ملفات الكود.

#### Scenario: ملف React/TypeScript
- **WHEN** فتح ملف مكون React
- **THEN** يجب أن يحتوي على تعليق رأسي يوضح الغرض من المكون

#### Scenario: ملف Python
- **WHEN** فتح ملف Python
- **THEN** يجب أن يحتوي على docstring رئيسي يوضح الوحدة

#### Scenario: ملف Go
- **WHEN** فتح ملف Go
- **THEN** يجب أن يحتوي على تعليق package يوضح الغرض

### Requirement: توحيد تسمية الملفات
The system SHALL يستخدم اتفاقية تسمية موحدة للملفات في كل قسم.

#### Scenario: مكونات React
- **WHEN** إنشاء مكون React جديد
- **THEN** يجب استخدام PascalCase (مثل: `GlassButton.tsx`)

#### Scenario: ملفات Python
- **WHEN** إنشاء ملف Python جديد
- **THEN** يجب استخدام snake_case (مثل: `event_store.py`)

#### Scenario: ملفات Go
- **WHEN** إنشاء ملف Go جديد
- **THEN** يجب استخدام snake_case (مثل: `config.go`)

### Requirement: إنشاء Style Guide
The system SHALL يوفر وثيقة Style Guide موحدة للمشروع.

#### Scenario: التوثيق
- **WHEN** مطور جديد ينضم للمشروع
- **THEN** يجب أن يجد وثيقة توضح معايير الكود المتبعة

## MODIFIED Requirements
لا توجد متطلبات معدلة لهذا التغيير.

## REMOVED Requirements
لا توجد متطلبات محذوفة لهذا التغيير.
