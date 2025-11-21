import React from 'react';

declare module '*.svg' {
    const content: React.FunctionComponent<React.SVGAttributes<SVGElement>>;
    export default content;
}

declare module 'mdui';
declare module 'mdui/dist/js/mdui.js';
declare module 'mdui/dist/css/mdui.css';
declare module 'countries-and-timezones' {
    const ct: any;
    export default ct;
}
declare module 'react-table' {
    const ReactTable: any;
    export default ReactTable;
}
