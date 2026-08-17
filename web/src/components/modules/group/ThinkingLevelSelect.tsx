'use client';

import type { ThinkingLevel } from '@/api/endpoints/group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useTranslations } from 'next-intl';

const INHERIT_VALUE = 'inherit';

function parseThinkingLevel(value: string): ThinkingLevel {
    switch (value) {
        case INHERIT_VALUE:
            return '';
        case 'minimal':
        case 'low':
        case 'medium':
        case 'high':
        case 'max':
            return value;
        default:
            return '';
    }
}

export function ThinkingLevelSelect({
    value,
    onChange,
}: {
    value: ThinkingLevel;
    onChange: (value: ThinkingLevel) => void;
}) {
    const t = useTranslations('group.thinkingLevel');

    return (
        <Select value={value || INHERIT_VALUE} onValueChange={(next) => onChange(parseThinkingLevel(next))}>
            <SelectTrigger size="sm" className="h-6 w-24 px-2 text-xs shadow-none" aria-label={t('label')}>
                <SelectValue />
            </SelectTrigger>
            <SelectContent align="end">
                <SelectItem value={INHERIT_VALUE}>{t('inherit')}</SelectItem>
                <SelectItem value="minimal">{t('minimal')}</SelectItem>
                <SelectItem value="low">{t('low')}</SelectItem>
                <SelectItem value="medium">{t('medium')}</SelectItem>
                <SelectItem value="high">{t('high')}</SelectItem>
                <SelectItem value="max">{t('max')}</SelectItem>
            </SelectContent>
        </Select>
    );
}
