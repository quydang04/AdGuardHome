import React, { useState, useEffect, useMemo } from 'react';

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

export const Form = ({
    initialValues,
    blockedServices,
    serviceGroups,
    processing,
    processingSet,
    onSubmit,
}: FormProps) => {
    const { t, i18n } = useTranslation();
    const [servicesLoaded, setServicesLoaded] = useState(true);

    useEffect(() => {
        setServicesLoaded(true);
    }, [i18n.language]);
    const {
        handleSubmit,
        control,
        setValue,
        formState: { isSubmitting },
    } = useForm<FormValues>({
        mode: 'onBlur',
        defaultValues: { blocked_services: initialValues }
    });

    const [masterEnabled] = useState<boolean>(true);

    const isBasicDisabled = processing || processingSet;
    const isSubmitDisabled = processing || processingSet || isSubmitting;

    const servicesByGroup = useMemo(() => {
        return blockedServices.reduce((acc, service) => {
            if (!acc[service.group_id]) {
                acc[service.group_id] = [];
            }
            acc[service.group_id].push(service);
            return acc;
        }, {} as Record<string, BlockedService[]>);
    }, [blockedServices]);

    const handleToggleAllServices = (isSelected: boolean) => {
        if (!masterEnabled) {
            return;
        }
        blockedServices.forEach((service) => {
            if (!isBasicDisabled) {
                setValue(`blocked_services.${service.id}`, isSelected);
            }
        });
    };

    const handleToggleGroupServices = (groupId: string, isSelected: boolean) => {
        if (isBasicDisabled || !masterEnabled) {
            return;
        }
        servicesByGroup[groupId].forEach((service) => {
            setValue(`blocked_services.${service.id}`, isSelected);
        });
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
                <div className="blocked_services row mb-3">
                    <div className="col-12 col-md-6 mb-4 mb-md-0">
                        <button
                            type="button"
                            data-testid="blocked_services_block_all"
                            className="btn btn-secondary btn-block font-weight-normal"
                            disabled={isBasicDisabled}
                            onClick={() => handleToggleAllServices(true)}>
                            <Trans>block_all</Trans>
                        </button>
                    </div>
                    <div className="col-12 col-md-6">
                        <button
                            type="button"
                            data-testid="blocked_services_unblock_all"
                            className="btn btn-secondary btn-block font-weight-normal"
                            disabled={isBasicDisabled}
                            onClick={() => handleToggleAllServices(false)}>
                            <Trans>unblock_all</Trans>
                        </button>
                    </div>
                </div>

                {servicesLoaded && (
                    <Accordion
                        items={serviceGroups.map((group) => {
                            return {
                                id: group.id,
                                title: t(`servicesgroup.${group.id}.name`, { ns: 'services' }),
                                disabled: processing || processingSet || !masterEnabled,
                            children: (
                                <div className="services__wrapper">
                                    <div className="row mb-3">
                                        <div className="col-12 col-md-6 mb-4 mb-md-0">
                                            <button
                                                type="button"
                                                className="btn btn-secondary btn-block font-weight-normal"
                                                disabled={isBasicDisabled}
                                                onClick={() => handleToggleGroupServices(group.id, true)}
                                            >
                                                <Trans>block_all</Trans>
                                            </button>
                                        </div>
                                        <div className="col-12 col-md-6">
                                            <button
                                                type="button"
                                                className="btn btn-secondary btn-block font-weight-normal"
                                                disabled={isBasicDisabled}
                                                onClick={() => handleToggleGroupServices(group.id, false)}
                                            >
                                                <Trans>unblock_all</Trans>
                                            </button>
                                        </div>
                                    </div>
                                    <div className="services">
                                        {servicesByGroup[group.id].map((service) => (
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
                                                            disabled={!masterEnabled || isBasicDisabled}
                                                            icon={service.icon_svg}
                                                        />
                                                )} />
                                            ))}
                                    </div>
                                </div>
                            ),
                            defaultOpen: true,
                        };
                    })}
                        allowMultiple
                        className="services-accordion"
                    />
                )}
            </div>

            <div className="btn-list">
                <button
                    type="submit"
                    data-testid="blocked_services_save"
                    className="btn btn-success btn-standard btn-large"
                    disabled={isSubmitDisabled}>
                    <Trans>save_btn</Trans>
                </button>
            </div>
        </form>
    );
};
