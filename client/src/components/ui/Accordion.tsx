import React, { useState } from 'react';
import './Accordion.css';
import '../Settings/Clients/Service.css';
import { AccordionItem } from './AccordionItem';

type AccordionProps = {
    items: any[];
    allowMultiple?: boolean;
    className?: string;
};

export const Accordion = ({
    items,
    allowMultiple = false,
    className = '',
}: AccordionProps) => {
    const [openItems, setOpenItems] = useState<Set<string>>(() => {
        const defaultOpen = new Set<string>();
        items.forEach((item) => {
            if (item.defaultOpen) {
                defaultOpen.add(item.id);
            }
        });
        return defaultOpen;
    });

    const toggleItem = (itemId: string) => {
        setOpenItems((prev) => {
            const newOpenItems = new Set(prev);

            if (newOpenItems.has(itemId)) {
                newOpenItems.delete(itemId);
            } else {
                if (!allowMultiple) {
                    newOpenItems.clear();
                }
                newOpenItems.add(itemId);
            }

            return newOpenItems;
        });
    };

    return (
        <div className={`accordion ${className}`}>
            {items.map((item) => (
                <AccordionItem
                    key={item.id}
                    id={item.id}
                    title={item.title}
                    isOpen={openItems.has(item.id)}
                    onToggle={() => toggleItem(item.id)}
                    disabled={item.disabled}
                >
                    {item.children}
                </AccordionItem>
            ))}
        </div>
    );
};

export default Accordion;
