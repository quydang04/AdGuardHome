import React, { useEffect, useMemo, useState } from 'react';

import { Trans, useTranslation } from 'react-i18next';

import { Controller, useForm } from 'react-hook-form';

import { ServiceField } from './ServiceField';
import { Accordion } from '../../ui/Accordion';

export type BlockedService = {
    id: string;
    name: string;
    icon_svg: string;
    group_id: string;
};

export type ServiceGroups = {
    id: string;
}

type FormValues = {
    blocked_services: Record<string, boolean>;
};

interface FormProps {
    initialValues: Record<string, boolean>;
    blockedServices: BlockedService[];
    serviceGroups: ServiceGroups[];
    onSubmit: (values: FormValues) => void;
    processing: boolean;
    processingSet: boolean;
}

const isServiceDisabled = (processing: boolean, processingSet: boolean) =>
  processing || processingSet;

export const Form = ({
    initialValues,
    blockedServices,
    serviceGroups,
    processing,
    processingSet,
    onSubmit,
}: FormProps) => {
    const { t } = useTranslation();
    const {
        handleSubmit,
        control,
        setValue,
        formState: { isSubmitting },
    } = useForm<FormValues>({
        mode: 'onBlur',
        defaultValues: { blocked_services: initialValues }
    });

    const [masterEnabled, setMasterEnabled] = useState<boolean>(true);

    const [groupEnabled, setGroupEnabled] = useState<Record<string, boolean>>(() =>
        serviceGroups.reduce<Record<string, boolean>>((acc, group) => {
            acc[group.id] = true;
            return acc;
        }, {})
    );

    useEffect(() => {
        setGroupEnabled(prev => {
            const missingGroups = serviceGroups.filter(group => !(group.id in prev));
            if (missingGroups.length === 0) {
                return prev;
            }

            const newGroups = Object.fromEntries(missingGroups.map(group => [group.id, true]));
            return { ...prev, ...newGroups };
        });
    }, [serviceGroups]);

    const groupToggleDisabled = useMemo(() => {
        return serviceGroups.reduce<Record<string, boolean>>(
            (groupDisabledMap, group) => {
                const servicesInGroup = blockedServices.filter(
                    (service) => service.group_id === group.id
                );

                const isGroupDisabled =
                    servicesInGroup.length > 0 &&
                    (isServiceDisabled(processing, processingSet) || !masterEnabled);

                return {
                    ...groupDisabledMap,
                    [group.id]: isGroupDisabled,
                };
            },
            {}
        );
    }, [serviceGroups, blockedServices, processing, processingSet, masterEnabled]);


    const handleToggleAllServices = async (isSelected: boolean) => {
        blockedServices.forEach((service) => {
            if (!isServiceDisabled(processing, processingSet)) {
                setValue(`blocked_services.${service.id}`, isSelected);
            }
        });
    };

    const handleGroupToggle = (groupId: string, enabled: boolean) => {
        if (groupToggleDisabled[groupId]) {
            return;
        }

        setGroupEnabled((prev) => ({ ...prev, [groupId]: enabled }));
    };

    const computedGroupStates = groupEnabled;

    const handleMasterToggle = (next: boolean) => {
        setMasterEnabled(next);
    };

    const handleSubmitWithGroups = (values: FormValues) => {
        if (!values || !values.blocked_services) {
            return onSubmit(values);
        }
        if (!masterEnabled) {
            return onSubmit({ blocked_services: {} });
        }

        const enabledIdsMap = Object.fromEntries(
            blockedServices
                .filter(service => values.blocked_services?.[service.id] && groupEnabled[service.group_id])
                .map(service => [service.id, true] as const)
        );

        return onSubmit({ blocked_services: enabledIdsMap });
    };

    return (
        <form onSubmit={handleSubmit(handleSubmitWithGroups)}>
            <div className="form__group">
                <ServiceField
                    name="blocked_services_master"
                    value={masterEnabled}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleMasterToggle(e.target.checked)}
                    onBlur={() => {}}
                    placeholder={t('blocked_services_global')}
                    className="service--global"
                    disabled={processing || processingSet}
                />
                <div className="row mb-4">
                    <div className="col-6">
                        <button
                            type="button"
                            data-testid="blocked_services_block_all"
                            className="btn btn-secondary btn-block"
                            disabled={processing || processingSet || !masterEnabled}
                            onClick={() => handleToggleAllServices(true)}>
                            <Trans>block_all</Trans>
                        </button>
                    </div>

                    <div className="col-6">
                        <button
                            type="button"
                            data-testid="blocked_services_unblock_all"
                            className="btn btn-secondary btn-block"
                            disabled={processing || processingSet || !masterEnabled}
                            onClick={() => handleToggleAllServices(false)}>
                            <Trans>unblock_all</Trans>
                        </button>
                    </div>
                </div>

                <Accordion
                    items={serviceGroups.map((group) => {
                        const disabled = groupToggleDisabled[group.id];

                        return {
                            id: group.id,
                            title: t(group.id),
                            disabled,
                            children: (
                                <div className={`services${disabled || !groupEnabled[group.id] ? ' is-group-disabled' : ''}`}>
                                    {blockedServices
                                        .filter((service) => service.group_id === group.id)
                                        .map((service) => (
                                            <Controller
                                                key={service.id}
                                                name={`blocked_services.${service.id}`}
                                                control={control}
                                                render={({ field }) => (
                                                    <ServiceField
                                                        {...field}
                                                        data-testid={`blocked_services_${service.id}`}
                                                        data-groupid={`blocked_services_${service.group_id}`}
                                                        placeholder={service.name}
                                                        disabled={
                                                            isServiceDisabled(processing, processingSet) ||
                                                            !masterEnabled ||
                                                            !groupEnabled[service.group_id]
                                                        }
                                                        icon={service.icon_svg}
                                                    />
                                                )}
                                            />
                                        ))}
                                </div>
                            ),
                            defaultOpen: true,
                        };
                    })}
                    allowMultiple
                    onGroupToggle={handleGroupToggle}
                    groupEnabledStates={computedGroupStates}
                    className="services-accordion"
                />
            </div>

            <div className="btn-list">
                <button
                    type="submit"
                    data-testid="blocked_services_save"
                    className="btn btn-success btn-standard btn-large"
                    disabled={isSubmitting || processing || processingSet}>
                    <Trans>save_btn</Trans>
                </button>
            </div>
        </form>
    );
};
