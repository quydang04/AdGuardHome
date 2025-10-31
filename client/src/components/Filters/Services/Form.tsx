import React, { useState } from 'react';

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

    // Group-level freeze switch removed; groups are always active. Buttons allow mass toggle.


    const handleToggleAllServices = async (isSelected: boolean) => {
        blockedServices.forEach((service) => {
            if (!isServiceDisabled(processing, processingSet)) {
                setValue(`blocked_services.${service.id}`, isSelected);
            }
        });
    };

    const handleToggleGroupServices = (groupId: string, isSelected: boolean) => {
        if (isServiceDisabled(processing, processingSet)) {
            return;
        }
        blockedServices
            .filter((s) => s.group_id === groupId)
            .forEach((service) => {
                setValue(`blocked_services.${service.id}`, isSelected);
            });
    };

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
                .filter(service => values.blocked_services?.[service.id])
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
                <div className="blocked_services row mb-4">
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
                        return {
                            id: group.id,
                            title: t(group.id),
                            children: (
                                <div className="services__wrapper">
                                    <div className="row mb-3">
                                        <div className="col-6">
                                            <button
                                                type="button"
                                                className="btn btn-secondary btn-block"
                                                disabled={processing || processingSet || !masterEnabled}
                                                onClick={() => handleToggleGroupServices(group.id, true)}
                                            >
                                                <Trans>block_all</Trans>
                                            </button>
                                        </div>
                                        <div className="col-6">
                                            <button
                                                type="button"
                                                className="btn btn-secondary btn-block"
                                                disabled={processing || processingSet || !masterEnabled}
                                                onClick={() => handleToggleGroupServices(group.id, false)}
                                            >
                                                <Trans>unblock_all</Trans>
                                            </button>
                                        </div>
                                    </div>
                                    <div className="services">
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
                                                            disabled={isServiceDisabled(processing, processingSet) || !masterEnabled}
                                                            icon={service.icon_svg} />
                                                    )} />
                                            ))}
                                    </div>
                                </div>
                            ),
                            defaultOpen: true,
                        };
                    })}
                    allowMultiple

                    className="services-accordion" onGroupToggle={undefined}                />
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
