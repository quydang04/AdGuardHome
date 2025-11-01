import React from 'react';

import { Trans, useTranslation } from 'react-i18next';

import { Controller, useForm } from 'react-hook-form';

import { ServiceField } from './ServiceField';
import { Accordion } from '../../ui/Accordion';
import { preloadServicesLocale } from '../../../helpers/servicesI18n';

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

const isServiceDisabled = (processing: boolean, processingSet: boolean) => processing || processingSet;

export const Form = ({
    initialValues,
    blockedServices,
    serviceGroups,
    processing,
    processingSet,
    onSubmit,
}: FormProps) => {
    const { t, i18n } = useTranslation();
    React.useEffect(() => {
        preloadServicesLocale(i18n.language).catch(() => {});
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

    const setAllServicesState = (isSelected: boolean) => {
        blockedServices.forEach((service) => {
            if (!isServiceDisabled(processing, processingSet)) {
                setValue(`blocked_services.${service.id}`, isSelected);
            }
        });
    };

    const setGroupServicesState = (groupId: string, isSelected: boolean) => {
        if (isServiceDisabled(processing, processingSet)) {
            return;
        }
        blockedServices
            .filter((s) => s.group_id === groupId)
            .forEach((service) => {
                setValue(`blocked_services.${service.id}`, isSelected);
            });
    };

    const handleSubmitWithGroups = (values: FormValues) => {
        if (!values || !values.blocked_services) {
            return onSubmit(values);
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
                <div className="blocked_services row mb-4">
                    <div className="col-6">
                        <button
                            type="button"
                            data-testid="blocked_services_block_all"
                            className="btn btn-secondary btn-block"
                            disabled={processing || processingSet}
                            onClick={() => setAllServicesState(true)}>
                            <Trans>block_all</Trans>
                        </button>
                    </div>

                    <div className="col-6">
                        <button
                            type="button"
                            data-testid="blocked_services_unblock_all"
                            className="btn btn-secondary btn-block"
                            disabled={processing || processingSet}
                            onClick={() => setAllServicesState(false)}>
                            <Trans>unblock_all</Trans>
                        </button>
                    </div>
                </div>

                <Accordion
                    items={serviceGroups.map((group) => {
                        return {
                            id: group.id,
                            title: t(`servicesgroup.${group.id}.name`),
                            children: (
                                <div className="services__wrapper">
                                    <div className="row mb-3">
                                        <div className="col-6">
                                            <button
                                                type="button"
                                                className="btn btn-secondary btn-block"
                                                disabled={processing || processingSet}
                                                onClick={() => setGroupServicesState(group.id, true)}
                                            >
                                                <Trans>block_all</Trans>
                                            </button>
                                        </div>
                                        <div className="col-6">
                                            <button
                                                type="button"
                                                className="btn btn-secondary btn-block"
                                                disabled={processing || processingSet}
                                                onClick={() => setGroupServicesState(group.id, false)}
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
                                                            disabled={isServiceDisabled(processing, processingSet)}
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
