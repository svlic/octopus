'use client';

import { useChannelList } from '@/api/endpoints/channel';
import { useStatsModel, type StatsModelFormatted } from '@/api/endpoints/stats';
import { useMemo } from 'react';
import { useTranslations } from 'next-intl';
import { TrendingUp } from 'lucide-react';
import { Tabs, TabsList, TabsTrigger, TabsContents, TabsContent } from '@/components/animate-ui/components/animate/tabs';
import { useHomeViewStore, type RankDimension, type RankSortMode } from '@/components/modules/home/store';
import { getModelIcon } from '@/lib/model-icons';

type ChannelData = NonNullable<ReturnType<typeof useChannelList>['data']>[number];

type RankItem = {
    key: string;
    name: string;
    totalCost: number;
    totalToken: number;
    requestCount: number;
    requestSuccess: number;
    requestFailed: number;
    totalCostFormatted: { value: string; unit: string };
    totalTokenFormatted: { value: string; unit: string };
    requestSuccessFormatted: { value: string; unit: string };
    requestFailedFormatted: { value: string; unit: string };
    showModelIcon?: boolean;
};

function channelToRankItem(channel: ChannelData): RankItem {
    return {
        key: `channel-${channel.raw.id}`,
        name: channel.raw.name,
        totalCost: channel.formatted.total_cost.raw,
        totalToken: channel.formatted.total_token.raw,
        requestCount: channel.formatted.request_count.raw,
        requestSuccess: channel.formatted.request_success.raw,
        requestFailed: channel.formatted.request_failed.raw,
        totalCostFormatted: channel.formatted.total_cost.formatted,
        totalTokenFormatted: channel.formatted.total_token.formatted,
        requestSuccessFormatted: channel.formatted.request_success.formatted,
        requestFailedFormatted: channel.formatted.request_failed.formatted,
    };
}

function modelToRankItem(model: StatsModelFormatted): RankItem {
    return {
        key: `model-${model.id || model.name}`,
        name: model.name,
        totalCost: model.total_cost.raw,
        totalToken: model.total_token.raw,
        requestCount: model.request_count.raw,
        requestSuccess: model.request_success.raw,
        requestFailed: model.request_failed.raw,
        totalCostFormatted: model.total_cost.formatted,
        totalTokenFormatted: model.total_token.formatted,
        requestSuccessFormatted: model.request_success.formatted,
        requestFailedFormatted: model.request_failed.formatted,
        showModelIcon: true,
    };
}

function sortRankItems(items: RankItem[], mode: RankSortMode): RankItem[] {
    const sorted = [...items];
    switch (mode) {
        case 'count':
            return sorted.sort((a, b) => b.requestCount - a.requestCount);
        case 'tokens':
            return sorted.sort((a, b) => b.totalToken - a.totalToken);
        case 'cost':
        default:
            return sorted.sort((a, b) => b.totalCost - a.totalCost);
    }
}

export function Rank() {
    const { data: channelData } = useChannelList();
    const { data: modelData } = useStatsModel();
    const t = useTranslations('home.rank');
    const rankSortMode = useHomeViewStore((state) => state.rankSortMode);
    const setRankSortMode = useHomeViewStore((state) => state.setRankSortMode);
    const rankDimension = useHomeViewStore((state) => state.rankDimension);
    const setRankDimension = useHomeViewStore((state) => state.setRankDimension);

    const channelItems = useMemo<RankItem[]>(() => {
        if (!channelData) return [];
        return channelData.map(channelToRankItem);
    }, [channelData]);

    const modelItems = useMemo<RankItem[]>(() => {
        if (!modelData) return [];
        return modelData.map(modelToRankItem);
    }, [modelData]);

    const sourceItems = rankDimension === 'model' ? modelItems : channelItems;

    const rankedByCost = useMemo(() => sortRankItems(sourceItems, 'cost'), [sourceItems]);
    const rankedByCount = useMemo(() => sortRankItems(sourceItems, 'count'), [sourceItems]);
    const rankedByTokens = useMemo(() => sortRankItems(sourceItems, 'tokens'), [sourceItems]);

    const getMedalEmoji = (rank: number): string => {
        switch (rank) {
            case 1: return '🥇';
            case 2: return '🥈';
            case 3: return '🥉';
            default: return '';
        }
    };

    const renderList = (items: RankItem[], mode: RankSortMode) => {
        if (items.length === 0) {
            return (
                <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                    <TrendingUp className="w-12 h-12 mb-3 opacity-30" />
                    <p className="text-sm">{t('noData')}</p>
                </div>
            );
        }
        return (
            <div className="space-y-3 max-h-[300px] overflow-y-auto">
                {items.map((item, index) => {
                    const rank = index + 1;
                    const medal = getMedalEmoji(rank);
                    const icon = item.showModelIcon ? getModelIcon(item.name) : null;
                    const ModelAvatar = icon?.Avatar;

                    return (
                        <div
                            key={item.key}
                            className="flex items-center gap-3 p-3 rounded-2xl hover:bg-accent/5 transition-colors"
                        >
                            <div className="w-8 h-8 rounded-lg flex items-center justify-center font-bold text-lg shrink-0">
                                {medal || rank}
                            </div>

                            {ModelAvatar ? (
                                <div className="w-8 h-8 rounded-lg overflow-hidden shrink-0 flex items-center justify-center bg-muted/40">
                                    <ModelAvatar size={28} />
                                </div>
                            ) : null}

                            <div className="flex-1 min-w-0">
                                <p className="font-medium text-sm truncate">{item.name}</p>
                                {mode === 'count' && (() => {
                                    const totalCount = item.requestSuccess + item.requestFailed;
                                    const successRate = totalCount > 0 ? (item.requestSuccess / totalCount) * 100 : 0;

                                    return (
                                        <div className="flex items-center gap-1 text-xs text-muted-foreground mt-1">
                                            <span>{t('successRate')}:</span>
                                            <span>{successRate.toFixed(1)}%</span>
                                        </div>
                                    );
                                })()}
                            </div>

                            <div className="flex items-center gap-1 text-right shrink-0">
                                {mode === 'count' ? (
                                    <div className="flex items-center gap-1 text-sm font-medium tabular-nums">
                                        <span className="text-accent">
                                            {item.requestSuccessFormatted.value}
                                            <span className="text-xs text-muted-foreground">
                                                {item.requestSuccessFormatted.unit}
                                            </span>
                                        </span>
                                        <span className="text-muted-foreground/40 font-light">/</span>
                                        <span className="text-destructive">
                                            {item.requestFailedFormatted.value}
                                            <span className="text-xs text-muted-foreground">
                                                {item.requestFailedFormatted.unit}
                                            </span>
                                        </span>
                                    </div>
                                ) : mode === 'tokens' ? (
                                    <span className="font-semibold text-base">
                                        {item.totalTokenFormatted.value}
                                        <span className="text-xs text-muted-foreground">
                                            {item.totalTokenFormatted.unit}
                                        </span>
                                    </span>
                                ) : (
                                    <span className="font-semibold text-base">
                                        {item.totalCostFormatted.value}
                                        <span className="text-xs text-muted-foreground">
                                            {item.totalCostFormatted.unit}
                                        </span>
                                    </span>
                                )}
                            </div>
                        </div>
                    );
                })}
            </div>
        );
    };

    return (
        <div className="rounded-3xl bg-card text-card-foreground border-card-border border p-4">
            <Tabs value={rankSortMode} onValueChange={(value) => setRankSortMode(value as RankSortMode)}>
                <div className="flex items-center justify-between gap-2 flex-wrap">
                    <div className="flex items-center gap-2 min-w-0">
                        <h3 className="font-semibold text-base shrink-0">{t('title')}</h3>
                        <Tabs
                            value={rankDimension}
                            onValueChange={(value) => setRankDimension(value as RankDimension)}
                        >
                            <TabsList>
                                <TabsTrigger value="channel">{t('dimensionChannel')}</TabsTrigger>
                                <TabsTrigger value="model">{t('dimensionModel')}</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </div>
                    <TabsList>
                        <TabsTrigger value="cost">{t('sortByCost')}</TabsTrigger>
                        <TabsTrigger value="count">{t('sortByCount')}</TabsTrigger>
                        <TabsTrigger value="tokens">{t('sortByTokens')}</TabsTrigger>
                    </TabsList>
                </div>
                <TabsContents>
                    <TabsContent value="cost">
                        {renderList(rankedByCost, 'cost')}
                    </TabsContent>
                    <TabsContent value="count">
                        {renderList(rankedByCount, 'count')}
                    </TabsContent>
                    <TabsContent value="tokens">
                        {renderList(rankedByTokens, 'tokens')}
                    </TabsContent>
                </TabsContents>
            </Tabs>
        </div>
    );
}
