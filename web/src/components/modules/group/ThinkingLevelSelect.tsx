'use client';

import { useEffect, useRef, useState } from 'react';
import { THINKING_LEVELS, type ThinkingLevel } from '@/api/endpoints/group';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useTranslations } from 'next-intl';

const CUSTOM_VALUE = 'custom';
const CUSTOM_VALUE_PATTERN = /^[A-Za-z0-9._-]*$/;

function isPreset(value: string): boolean {
    return THINKING_LEVELS.some((level) => level === value);
}

export function ThinkingLevelSelect({
    value,
    onChange,
}: {
    value: ThinkingLevel;
    onChange: (value: ThinkingLevel) => void;
}) {
    const t = useTranslations('group.thinkingLevel');
    const normalizedValue = value || 'default';
    const initialCustomActive = !isPreset(normalizedValue);
    const [customDraft, setCustomDraft] = useState<ThinkingLevel | null>(initialCustomActive ? value : null);
    const lastEmittedValue = useRef<ThinkingLevel>(value);

    useEffect(() => {
        if (value === lastEmittedValue.current) return;
        const timer = window.setTimeout(() => {
            lastEmittedValue.current = value;
            setCustomDraft(isPreset(value || 'default') ? null : value);
        });
        return () => window.clearTimeout(timer);
    }, [value]);

    const emitChange = (next: ThinkingLevel) => {
        lastEmittedValue.current = next;
        onChange(next);
    };
    const customActive = customDraft !== null;
    const customValue = customDraft ?? '';
    const selectValue = customActive ? CUSTOM_VALUE : normalizedValue;

    return (
        <div className="flex shrink-0 items-center gap-1.5">
            <Select
                value={selectValue}
                onValueChange={(next) => {
                    if (next === CUSTOM_VALUE) {
                        setCustomDraft('');
                        emitChange('');
                        return;
                    }
                    setCustomDraft(null);
                    emitChange(next);
                }}
            >
                <SelectTrigger size="sm" className="h-7 w-24 px-2 font-mono text-xs shadow-none" aria-label={t('label')}>
                    <SelectValue />
                </SelectTrigger>
                <SelectContent align="end">
                    {THINKING_LEVELS.map((level) => (
                        <SelectItem key={level} value={level}>{level}</SelectItem>
                    ))}
                    <SelectItem value={CUSTOM_VALUE}>{CUSTOM_VALUE}</SelectItem>
                </SelectContent>
            </Select>
            {customActive && (
                <Input
                    value={customValue}
                    onChange={(event) => {
                        const next = event.target.value;
                        if (CUSTOM_VALUE_PATTERN.test(next)) {
                            setCustomDraft(next);
                            emitChange(next);
                        }
                    }}
                    onBlur={() => {
                        if (!customValue) {
                            setCustomDraft(null);
                            emitChange('default');
                        }
                    }}
                    maxLength={64}
                    inputMode="text"
                    autoCapitalize="none"
                    spellCheck={false}
                    aria-label="custom thinking level"
                    placeholder="custom"
                    className="h-7 w-28 rounded-md px-2 font-mono text-xs shadow-none"
                />
            )}
        </div>
    );
}
