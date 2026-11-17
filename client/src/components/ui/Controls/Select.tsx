import React, { ComponentProps, forwardRef } from 'react';
import clsx from 'clsx';

type SelectProps = ComponentProps<'select'> & {
    label?: string;
    error?: string;
};

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
    ({ id, name, label, className, error, children, ...rest }, ref) => (
        <div className={clsx('form-group', { 'has-error': !!error })}>
            {label && (
                <label className="form__label" htmlFor={id || name}>
                    {label}
                </label>
            )}
            <div className="input-group">
                <select id={id || name} name={name} className={clsx('form-control custom-select', className)} ref={ref} {...rest}>
                    {children}
                </select>
            </div>
            {error && <div className="form__message form__message--error mt-1">{error}</div>}
        </div>
    ),
);

Select.displayName = 'Select';
