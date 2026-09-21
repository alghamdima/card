import { translate, type Locale } from '../i18n';
import type { TextFieldConfig } from '../types/campaign.types';

/** Default text fields for a new template, labelled in the template's own language. */
export function defaultFields(lang: Locale): TextFieldConfig[] {
  return [
    {
      id: 'emp_name',
      name: 'emp_name',
      label: translate(lang, 'admin.defaultNameLabel'),
      placeholder: translate(lang, 'admin.defaultNamePlaceholder'),
      x: 230,
      y: 620,
      width: 620,
      height: 70,
      fontSize: 47,
      color: '#FFFFFF',
      weight: 'bold',
      align: 'center',
      order: 1
    },
    {
      id: 'job_title',
      name: 'job_title',
      label: translate(lang, 'admin.defaultTitleLabel'),
      placeholder: translate(lang, 'admin.defaultTitlePlaceholder'),
      x: 230,
      y: 705,
      width: 620,
      height: 60,
      fontSize: 34,
      color: '#B9B9C2',
      weight: 'regular',
      align: 'center',
      order: 2
    }
  ];
}

/** Sample values shown in the builder preview. */
export function defaultSamples(lang: Locale): Record<string, string> {
  return {
    emp_name: translate(lang, 'admin.defaultNameSample'),
    job_title: translate(lang, 'admin.defaultTitleSample')
  };
}

export function newField(lang: Locale, existingCount: number): TextFieldConfig {
  const id = `field_${Date.now().toString(36)}`;
  return {
    id,
    name: id,
    label: translate(lang, 'admin.newFieldLabel', { n: existingCount + 1 }),
    x: 250,
    y: 780 + existingCount * 30,
    width: 580,
    height: 65,
    fontSize: 36,
    color: '#FFCD00',
    weight: 'bold',
    align: 'center',
    order: existingCount + 1
  };
}
