// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';
import React, { useEffect, useState } from 'react';
import { Tag } from 'antd';


const AllTagsList = ({ tags, handleClose, closable }) => {
    const tagsToList = () => {
        const tagsToMakeList = tags;
        const listItems = tagsToMakeList.map((tag) => {
            const isLongTag = tag.length > 20;
            <li>
                <Tag className="tag-wrapper" key={tag.name} color={tag.color} closable={closable ? closable : false} onClose={() => handleClose(tag.name)}>
                    {isLongTag ? `${tag.name.slice(0, 20)}...` : tag.name}
                </Tag>
            </li>;
        });
        return <ul>{listItems}</ul>;
    };
    return <>{tagsToList}</>;
};

export default AllTagsList;
