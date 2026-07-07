import React from "react";
import { Link } from "react-router-dom";
import { EmptyState } from "../components";
import { IconServices } from "../icons";

export function NotFound() {
  return (
    <EmptyState
      icon={<IconServices />}
      title="Page not found"
      description="The page you're looking for doesn't exist."
      action={
        <Link to="/" className="btn btn-primary">
          Back to catalog
        </Link>
      }
    />
  );
}
